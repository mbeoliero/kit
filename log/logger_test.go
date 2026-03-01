package log

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func TestInfo(t *testing.T) {
	// Initialize OpenTelemetry TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("test-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Now create tracer and span
	tracer := otel.Tracer("your-service")
	ctx, span := tracer.Start(context.Background(), "operation")
	defer span.End()

	SetProdEnv()
	// Test with trace ID
	CtxInfo(ctx, "no extra")
	ctx = AppendLogKv(ctx, "user_id", "1234")
	CtxInfo(ctx, "test %s", "abc")

	// Test without context (should show "-" for trace ID)
	Info("test %s", "abc")

	ctx = AppendLogExtras(ctx, map[string]string{"user_id": "aaa", "app_id": "2"})
	CtxInfo(ctx, "test app_id %s", "abc")
	Info("test app_id %s", "abc")
}

func TestInfoWithLogrus(t *testing.T) {
	// Test with logrus
	SetLoggerType(LoggerTypeLogrus)
	logger = newLogger()
	logger.SetLevel(klog.LevelDebug)
	defaultLogger = logger

	// Initialize OpenTelemetry
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("test-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("your-service")
	ctx, span := tracer.Start(context.Background(), "operation")
	defer span.End()

	CtxInfo(ctx, "logrus test with trace ID")
	Info("logrus test without trace ID")
}

func TestZerolog(t *testing.T) {
	cw := zerolog.NewConsoleWriter()

	// Create zerolog logger with proper configuration
	zlog := zerolog.New(cw).
		With().Timestamp().Logger().
		Hook(customFieldsHook{})

	// Use CallerWithSkipFrameCount to get correct caller location
	// Skip 5 frames to get to the actual user code
	zlog = zlog.With().CallerWithSkipFrameCount(5).Logger()
	zlog.Printf("zerolog test with trace ID %s", "1234")
}

func TestLoggerLevel(t *testing.T) {
	CtxDebug(context.Background(), "no extra test")
	SetProdEnv()
	//SetLoggerType(LoggerTypeLogrus)
	CtxDebug(context.Background(), "no extra prod")
}

func TestSetLevelInfoSuppressesDebugForZerolog(t *testing.T) {
	oldStdout := os.Stdout
	oldLogger := logger
	oldDefaultLogger := defaultLogger
	oldLogLevel := logLevel
	oldLoggerType := currentLoggerType
	oldCustomOut := customOut

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = w

	SetLoggerType(LoggerTypeZerolog)
	logger = newLogger()
	defaultLogger = logger

	defer func() {
		_ = w.Close()
		_ = r.Close()
		os.Stdout = oldStdout
		logger = oldLogger
		defaultLogger = oldDefaultLogger
		logLevel = oldLogLevel
		currentLoggerType = oldLoggerType
		customOut = oldCustomOut
	}()

	logger.SetLevel(klog.LevelInfo)
	Debug("debug should be filtered")
	Info("info should be logged")

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	bs, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read logs: %v", err)
	}
	output := string(bs)
	if strings.Contains(output, "debug should be filtered") {
		t.Fatalf("debug log should be filtered at info level, got output: %s", output)
	}
	if !strings.Contains(output, "info should be logged") {
		t.Fatalf("info log should be present, got output: %s", output)
	}
}
