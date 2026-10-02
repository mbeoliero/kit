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

const pingTimeout = 5 * time.Second

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
		log.Error("init redis %s failed: %v", redisCfg.Addr, err)
		panic(err)
	}
	return client
}

func MustInitClusterRedis(redisCfg RedisConfig) *redis.ClusterClient {
	client, err := InitClusterRedis(redisCfg)
	if err != nil {
		log.Error("init cluster redis %s failed: %v", redisCfg.Addr, err)
		panic(err)
	}
	return client
}

func InitRedis(redisCfg RedisConfig) (*redis.Client, error) {
	log.Info("init redis addr=%s db=%d pool=%d tls=%v", redisCfg.Addr, redisCfg.DB, redisCfg.PoolSize, redisCfg.EnableTLS)
	client := redis.NewClient(&redis.Options{
		Addr:      redisCfg.Addr,
		Username:  redisCfg.Username,
		Password:  redisCfg.Password,
		DB:        redisCfg.DB,
		PoolSize:  redisCfg.PoolSize,
		TLSConfig: redisTLS(redisCfg),
	})
	if err := setupRedis(client, redisCfg); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func InitClusterRedis(redisCfg RedisConfig) (*redis.ClusterClient, error) {
	log.Info("init cluster redis addr=%s pool=%d tls=%v master_only=%v", redisCfg.Addr, redisCfg.PoolSize, redisCfg.EnableTLS, redisCfg.MasterOnly)
	options := &redis.ClusterOptions{
		Addrs:     []string{redisCfg.Addr},
		Username:  redisCfg.Username,
		Password:  redisCfg.Password,
		PoolSize:  redisCfg.PoolSize,
		TLSConfig: redisTLS(redisCfg),
	}
	if !redisCfg.MasterOnly {
		options.ReadOnly = true
		options.RouteRandomly = true
	}
	client := redis.NewClusterClient(options)
	if err := setupRedis(client, redisCfg); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

// redisTLS verifies the server certificate unless the config explicitly opts out.
func redisTLS(cfg RedisConfig) *tls.Config {
	if !cfg.EnableTLS {
		return nil
	}
	return &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: cfg.TLSInsecureSkipVerify}
}

func setupRedis(client redis.UniversalClient, cfg RedisConfig) error {
	if cfg.EnableLog {
		client.AddHook(RedisHook{enableLog: true})
	}
	if !cfg.DisableTrace {
		if err := redisotel.InstrumentTracing(client); err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	return client.Ping(ctx).Err()
}

// RedisHook logs every command at debug level when enabled.
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
		return err
	}
}

func (r RedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cs []redis.Cmder) error {
		begin := time.Now()
		err := next(ctx, cs)
		if r.enableLog {
			cmdList := make([]string, 0, len(cs))
			for _, cmd := range cs {
				cmdList = append(cmdList, cmd.String())
			}
			log.CtxDebug(ctx, "[Redis Cmd][%v] %s", time.Since(begin), strings.Join(cmdList, ", "))
		}
		return err
	}
}
