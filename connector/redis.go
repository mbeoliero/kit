package connector

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/mbeoliero/kit/log"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func MustInitRedis(cfg RedisConfig) redis.UniversalClient {
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 1000
	}

	if cfg.IsCluster {
		return MustInitClusterRedis(cfg)
	}
	return MustInitDefaultRedis(cfg)
}

func MustInitDefaultRedis(redisCfg RedisConfig) *redis.Client {
	client, err := InitRedis(redisCfg)
	if err != nil {
		log.Error("init redis failed with error %v, cfg %+v", err, redisCfg)
		panic(err)
	}

	_, err = client.Ping(context.TODO()).Result()
	if err != nil {
		log.Error("ping redis failed with error %v, cfg %+v", err, redisCfg)
		panic(err)
	}
	return client
}

func MustInitClusterRedis(redisCfg RedisConfig) *redis.ClusterClient {
	client, err := InitClusterRedis(redisCfg)
	if err != nil {
		log.Error("init redis failed with error %v, cfg %+v", err, redisCfg)
		panic(err)
	}

	_, err = client.Ping(context.TODO()).Result()
	if err != nil {
		log.Error("ping redis failed with error %v, cfg %+v", err, redisCfg)
		panic(err)
	}
	return client
}

func InitRedis(redisCfg RedisConfig) (client *redis.Client, err error) {
	log.Info("init redis cfg=%+v", redisCfg)
	options := &redis.Options{
		Addr:     redisCfg.Addr,
		Username: redisCfg.Username,
		Password: redisCfg.Password, // no password set
		DB:       redisCfg.DB,       // use default DB
		PoolSize: redisCfg.PoolSize,
	}
	if redisCfg.EnableTLS {
		options.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	// 国内(腾讯)不支持3的协议，所以使用2的协议
	//if idc.IsCN() {
	//	options.Protocol = 2
	//}
	client = redis.NewClient(options)
	log.Info("init redis new client done")
	if err = injectRedisTracing(!redisCfg.DisableTrace, redisCfg.EnableLog, client); err != nil {
		return nil, err
	}

	log.Info("init redis inject redis trace done")
	_, err = client.Ping(context.TODO()).Result()
	log.Info("init redis ping done")
	if err != nil {
		return nil, err
	}
	return client, nil
}

func InitClusterRedis(redisCfg RedisConfig) (client *redis.ClusterClient, err error) {
	log.Info("init cluster redis cfg=%+v", redisCfg)
	options := &redis.ClusterOptions{
		Addrs:    []string{redisCfg.Addr},
		Username: redisCfg.Username,
		Password: redisCfg.Password, // no password set
		PoolSize: redisCfg.PoolSize,
	}
	// 国内(腾讯)不支持3的协议，所以使用2的协议
	//if idc.IsCN() {
	//	options.Protocol = 2
	//}
	if redisCfg.EnableTLS {
		options.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if !redisCfg.MasterOnly {
		options.ReadOnly = true
		options.RouteRandomly = true
		log.CtxInfo(context.TODO(), "init cluster redis set read only and route randomly")
	} else {
		log.CtxInfo(context.TODO(), "init cluster redis master only")
	}
	client = redis.NewClusterClient(options)
	log.Info("init cluster redis new client done")
	if err = injectRedisTracing(!redisCfg.DisableTrace, redisCfg.EnableLog, client); err != nil {
		return nil, err
	}

	log.Info("init cluster redis inject redis trace done")
	_, err = client.Ping(context.TODO()).Result()
	log.Info("init cluster redis ping done")
	if err != nil {
		return nil, err
	}
	return client, nil
}

func injectRedisTracing(enableTracing bool, enableLog bool, client redis.UniversalClient) error {
	if enableTracing {
		client.AddHook(RedisHook{enableLog: enableLog})
		return redisotel.InstrumentTracing(client)
	}
	return nil
}

type RedisHook struct {
	enableLog bool
}

var _ redis.Hook = RedisHook{}

func (RedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}
func (r RedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		begin := time.Now()
		err := next(ctx, cmd)

		if r.enableLog {
			log.CtxDebug(ctx, "[Redis Cmd][%v] %s", time.Since(begin), cmd.String())
		}

		addDbMetrics(redisDb, time.Now().Sub(begin).Milliseconds(), err)
		return err
	}
}
func (r RedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cs []redis.Cmder) error {
		begin := time.Now()
		err := next(ctx, cs)

		if r.enableLog {
			var cmdList []string
			for _, cmd := range cs {
				cmdList = append(cmdList, cmd.String())
			}
			log.CtxDebug(ctx, "[Redis Cmd][%v] %s", time.Since(begin), strings.Join(cmdList, ", "))
		}
		
		addDbMetrics(redisDb, time.Now().Sub(begin).Milliseconds(), err)
		return err
	}
}
