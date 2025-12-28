package log

import (
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// Ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/README.md#json-formats
const (
	//traceIDKey    = "trace_id"
	spanIDKey     = "span_id"
	traceFlagsKey = "trace_flags"
)

var _ logrus.Hook = (*traceIdHook)(nil)

type traceIdHook struct {
}

func (h *traceIdHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *traceIdHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}

	span := trace.SpanFromContext(entry.Context)

	// attach span context to log entry data fields
	entry.Data[TraceIDKey] = span.SpanContext().TraceID()
	//entry.Data[spanIDKey] = span.SpanContext().SpanID()
	//entry.Data[traceFlagsKey] = span.SpanContext().TraceFlags()

	return nil
}
