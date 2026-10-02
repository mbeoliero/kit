package connector

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/mbeoliero/kit/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const secret = "p@ss:/?#word"

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })
	return &buf
}

func TestInitLogsOmitCredentials(t *testing.T) {
	out := captureLogs(t)

	_, _ = InitGorm(MysqlConfig{Path: "127.0.0.1:1", Dbname: "db", Username: "u", Password: secret})
	_, _ = InitGorm(MysqlConfig{WritePath: "127.0.0.1:1", ReadPath: "127.0.0.1:2", Dbname: "db", Username: "u", Password: secret})
	_, _ = InitRedis(RedisConfig{Addr: "127.0.0.1:1", Password: secret})
	_, _ = InitMongo(MongoConfig{Address: "127.0.0.1:1", Username: "u", Password: secret, Cfg: "serverSelectionTimeoutMS=100&connectTimeoutMS=100"})

	if strings.Contains(out.String(), secret) || strings.Contains(out.String(), url.QueryEscape(secret)) {
		t.Fatalf("logs contain the password:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "127.0.0.1:1") {
		t.Fatalf("logs should still name the address:\n%s", out.String())
	}
}

func TestMongoUriEscapesCredentials(t *testing.T) {
	uri := mongoUri(MongoConfig{Address: "h1:27017,h2:27017", Username: "u", Password: secret, Cfg: "authSource=admin"})
	parsed, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	if password, _ := parsed.User.Password(); password != secret || parsed.Host != "h1:27017,h2:27017" || parsed.RawQuery != "authSource=admin" {
		t.Fatalf("uri = %s", uri)
	}
	if got := mongoUri(MongoConfig{Address: "localhost:27017"}); got != "mongodb://localhost:27017/" {
		t.Fatalf("anonymous uri = %s", got)
	}
}

func TestMongoMonitorKeepsTracing(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	original := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	t.Cleanup(func() { otel.SetTracerProvider(original) })
	out := captureLogs(t)
	log.SetLevel(log.LevelDebug)

	monitor := newMongoMonitor(true, false)
	command, _ := bson.Marshal(bson.D{{Key: "find", Value: "users"}})
	reply, _ := bson.Marshal(bson.D{{Key: "cursor", Value: bson.D{{Key: "firstBatch", Value: bson.A{bson.D{}, bson.D{}}}}}})
	ctx := context.Background()
	monitor.Started(ctx, &event.CommandStartedEvent{Command: command, DatabaseName: "db", CommandName: "find", RequestID: 7, ConnectionID: "localhost:27017[-1]"})
	monitor.Succeeded(ctx, &event.CommandSucceededEvent{
		CommandFinishedEvent: event.CommandFinishedEvent{CommandName: "find", DatabaseName: "db", RequestID: 7, ConnectionID: "localhost:27017[-1]"},
		Reply:                reply,
	})

	if n := len(recorder.Ended()); n != 1 {
		t.Fatalf("ended spans = %d, want 1", n)
	}
	if !strings.Contains(out.String(), "affected: 2") {
		t.Fatalf("affected count not logged:\n%s", out.String())
	}
}

func TestRedisTLSVerifiesByDefault(t *testing.T) {
	if redisTLS(RedisConfig{}) != nil {
		t.Fatal("TLS should stay off unless enabled")
	}
	if cfg := redisTLS(RedisConfig{EnableTLS: true}); cfg == nil || cfg.InsecureSkipVerify {
		t.Fatalf("tls = %+v, want verification", cfg)
	}
	if cfg := redisTLS(RedisConfig{EnableTLS: true, TLSInsecureSkipVerify: true}); !cfg.InsecureSkipVerify {
		t.Fatal("explicit opt-out ignored")
	}
}
