package log

import "context"

// Sink receives every record the built-in logger writes, after level filtering and
// with the message already formatted. Emit runs synchronously on the logging
// goroutine, so implementations should hand off slow work.
type Sink interface {
	Emit(ctx context.Context, level Level, msg string)
}

// AddSink registers an extra destination for the built-in logger, e.g. a log exporter.
func AddSink(s Sink) {
	builtin.addSink(s)
}
