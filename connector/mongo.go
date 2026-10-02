package connector

import (
	"context"
	"crypto/tls"
	"net/url"
	"time"

	"github.com/mbeoliero/kit/log"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
)

// slowMongoCommand is logged at warn level even when command logging is disabled.
const slowMongoCommand = time.Second

func MustInitMongo(cfg MongoConfig) (*mongo.Client, *mongo.Database) {
	cli, err := InitMongo(cfg)
	if err != nil {
		log.Error("init mongodb %s failed: %v", cfg.Address, err)
		panic(err)
	}
	return cli, cli.Database(cfg.Database)
}

func InitMongo(mgoCfg MongoConfig) (*mongo.Client, error) {
	log.Info("init mongo address=%s database=%s tls=%v", mgoCfg.Address, mgoCfg.Database, mgoCfg.EnableTLS)
	opt := options.Client().ApplyURI(mongoUri(mgoCfg))
	if mgoCfg.EnableTLS {
		opt.SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})
	}
	opt.SetMonitor(newMongoMonitor(!mgoCfg.DisableTrace, mgoCfg.DisableLog))

	cli, err := mongo.Connect(opt)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err = cli.Ping(ctx, readpref.Primary()); err != nil {
		_ = cli.Disconnect(context.Background())
		return nil, err
	}
	log.Info("init mongo done")
	return cli, nil
}

// mongoUri escapes credentials, so passwords may contain URI delimiters.
func mongoUri(cfg MongoConfig) string {
	u := url.URL{Scheme: "mongodb", Host: cfg.Address, Path: "/", RawQuery: cfg.Cfg}
	if cfg.Username != "" {
		u.User = url.UserPassword(cfg.Username, cfg.Password)
	}
	return u.String()
}

// newMongoMonitor chains command logging after the otelmongo monitor, so tracing spans
// still start and end.
func newMongoMonitor(enableTracing, disableLog bool) *event.CommandMonitor {
	tracing := &event.CommandMonitor{}
	if enableTracing {
		tracing = otelmongo.NewMonitor()
	}
	return &event.CommandMonitor{
		Started: func(ctx context.Context, e *event.CommandStartedEvent) {
			if tracing.Started != nil {
				tracing.Started(ctx, e)
			}
			if !disableLog {
				log.CtxDebug(ctx, "[Mongo Cmd] %s: %s", e.CommandName, e.Command.String())
			}
		},
		Succeeded: func(ctx context.Context, e *event.CommandSucceededEvent) {
			if tracing.Succeeded != nil {
				tracing.Succeeded(ctx, e)
			}
			if e.Duration >= slowMongoCommand {
				log.CtxWarn(ctx, "[Mongo Slow] cmd: %s, duration: %dms, affected: %d", e.CommandName, e.Duration.Milliseconds(), affectedCount(e))
			} else if !disableLog {
				log.CtxDebug(ctx, "[Mongo Succeeded] cmd: %s, duration: %dms, affected: %d", e.CommandName, e.Duration.Milliseconds(), affectedCount(e))
			}
		},
		Failed: func(ctx context.Context, e *event.CommandFailedEvent) {
			if tracing.Failed != nil {
				tracing.Failed(ctx, e)
			}
			log.CtxWarn(ctx, "[Mongo Failed] cmd: %s, duration: %dms, err: %v", e.CommandName, e.Duration.Milliseconds(), e.Failure)
		},
	}
}

func affectedCount(e *event.CommandSucceededEvent) int {
	switch e.CommandName {
	case "find":
		if docs, ok := e.Reply.Lookup("cursor", "firstBatch").ArrayOK(); ok {
			values, _ := docs.Values()
			return len(values)
		}
	case "update":
		if n, ok := e.Reply.Lookup("nModified").Int32OK(); ok {
			return int(n)
		}
	case "insert", "delete":
		if n, ok := e.Reply.Lookup("n").Int32OK(); ok {
			return int(n)
		}
	}
	return 0
}
