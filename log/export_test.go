package log

import (
	"bytes"
	"os"
	"testing"
)

// CaptureForTest redirects the built-in logger to a buffer and restores the
// package defaults when the test ends.
func CaptureForTest(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	builtin.SetOutput(&buf)
	t.Cleanup(func() {
		builtin.SetOutput(os.Stdout)
		builtin.sinks.Store(nil)
		defaultLogger = logger
		SetLevel(LevelDebug)
	})
	return &buf
}
