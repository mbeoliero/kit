package connector

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mbeoliero/kit/log"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
)

func MustInitMongo(cfg MongoConfig) (*mongo.Client, *mongo.Database) {
	cli, err := InitMongo(cfg)
	if err != nil {
		log.Error("init mongodb failed with error %v and cfg %+v", err, cfg)
		panic(err)
	}

	database := cli.Database(cfg.Database)
	return cli, database

}

func InitMongo(mgoCfg MongoConfig) (*mongo.Client, error) {
	url := fmt.Sprintf("mongodb://%s:%s@%s", mgoCfg.Username, mgoCfg.Password, mgoCfg.Address)
	if os.Getenv("") == "" && strings.Contains(url, "localhost") && mgoCfg.Password == "" {
		url = fmt.Sprintf("mongodb://%s", mgoCfg.Address)
	}
	if cfgStr := mgoCfg.Cfg; cfgStr != "" {
		url = fmt.Sprintf("%s/?%s", url, cfgStr)
	}

	opt := options.Client()
	injectMongoTracing(!mgoCfg.DisableTrace, mgoCfg.DisableLog, opt)

	log.Info("init mongo idle time= %d cfg=%+v", 10*time.Second, mgoCfg)
	cli, err := mongo.Connect(opt.ApplyURI(url))
	if err != nil {
		return nil, err
	}

	err = cli.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		return nil, err
	}
	log.Info("init mongo done")
	return cli, err
}

func injectMongoTracing(enableTracing bool, disableLog bool, clientOpt *options.ClientOptions) {
	if !enableTracing {
		return
	}
	clientOpt.Monitor = otelmongo.NewMonitor()
	clientOpt.Monitor.Started = func(ctx context.Context, event *event.CommandStartedEvent) {
		printSql(ctx, disableLog, event)
	}

	clientOpt.Monitor.Succeeded = func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
		var effectedCount int
		switch succeededEvent.CommandName {
		case "find":
			if nReturned, ok := succeededEvent.Reply.Lookup("cursor", "firstBatch").ArrayOK(); ok {
				effectedCount = len(nReturned)
			}

		case "insert":
			if nInserted, ok := succeededEvent.Reply.Lookup("n").Int32OK(); ok {
				effectedCount = int(nInserted)
			}

		case "update":
			if nModified, ok := succeededEvent.Reply.Lookup("nModified").Int32OK(); ok {
				effectedCount = int(nModified)
			}

		case "delete":
			if nDeleted, ok := succeededEvent.Reply.Lookup("n").Int32OK(); ok {
				effectedCount = int(nDeleted)
			}
		}

		ms := succeededEvent.Duration.Milliseconds()
		if ms > 1_000 {
			log.CtxWarn(ctx, "[Mongo Succeeded] cmd: %s, duration: %dms, effectedCount: %d", succeededEvent.CommandName, ms, effectedCount)
		} else {
			log.CtxDebug(ctx, "[Mongo Succeeded] cmd: %s, duration: %dms, effectedCount: %d", succeededEvent.CommandName, ms, effectedCount)
		}
		addDbMetrics(mongoDb, succeededEvent.Duration.Milliseconds(), nil)
	}

	clientOpt.Monitor.Failed = func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
		log.CtxError(ctx, "[Mongo Failed] cmd: %s, duration: %dms, err: %s", failedEvent.CommandName, failedEvent.Duration.Milliseconds(), failedEvent.Failure)
		addDbMetrics(mongoDb, failedEvent.Duration.Milliseconds(), errors.New("failedEvent.Failure"))
	}
}

func printSql(ctx context.Context, disableLog bool, event *event.CommandStartedEvent) {
	if disableLog {
		log.CtxDebug(ctx, "[Mongo Sql] sql %v: %+v", event.CommandName, event.Command.String())
		return
	}
	log.CtxInfo(ctx, "[Mongo Sql] sql %v: %+v", event.CommandName, event.Command.String())
}
