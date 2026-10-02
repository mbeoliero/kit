package log_test

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/mbeoliero/kit/log"
)

var linePattern = regexp.MustCompile(`^(\d{4}-\d\d-\d\d \d\d:\d\d:\d\d\.\d{3}) (\w+) (\d+) (\d+) (\S+) (\S+) (\{.*\}) : (.*)$`)

type record struct {
	time, level, traceId, caller, fields, msg string
}

func parseLines(t *testing.T, out string) []record {
	t.Helper()
	var records []record
	for line := range strings.SplitSeq(strings.TrimSuffix(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		m := linePattern.FindStringSubmatch(line)
		if m == nil {
			t.Fatalf("line does not match format: %q", line)
		}
		records = append(records, record{time: m[1], level: m[2], traceId: m[5], caller: m[6], fields: m[7], msg: m[8]})
	}
	return records
}

func here(offset int) string {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Sprintf("%s:%d", file, line+offset)
}

func TestLineFormat(t *testing.T) {
	out := log.CaptureForTest(t)
	tp := sdktrace.NewTracerProvider()
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	ctx, span := tp.Tracer("test").Start(t.Context(), "op")
	defer span.End()
	ctx = log.AppendLogExtras(ctx, map[string]string{"user_id": "1", "app": "a"})

	before := time.Now()
	caller := here(1)
	log.CtxInfo(ctx, "hello %d", 1)
	log.Info("plain")

	records := parseLines(t, out.String())
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	got := records[0]
	if got.level != "INFO" || got.msg != "hello 1" || got.caller != caller {
		t.Fatalf("record = %+v, want INFO hello 1 at %s", got, caller)
	}
	if want := span.SpanContext().TraceID().String(); got.traceId != want {
		t.Fatalf("trace id = %s, want %s", got.traceId, want)
	}
	if got.fields != `{"app":"a","user_id":"1"}` {
		t.Fatalf("fields = %s", got.fields)
	}
	if records[1].traceId != "-" || records[1].fields != "{}" {
		t.Fatalf("context-free record = %+v", records[1])
	}
	logged, err := time.ParseInLocation("2006-01-02 15:04:05.000", got.time, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	// Second-precision timestamps would land before the call.
	if logged.Before(before.Truncate(time.Millisecond)) {
		t.Fatalf("timestamp %s precedes call at %s", got.time, before.Format(time.StampMilli))
	}
}

func TestLevels(t *testing.T) {
	out := log.CaptureForTest(t)
	log.SetLevel(log.LevelDebug)
	log.Trace("trace")
	log.Notice("notice")
	log.SetLevel(log.LevelInfo)
	log.Debug("hidden")
	log.CtxDebug(t.Context(), "hidden")
	log.Info("shown")

	var got []string
	for _, r := range parseLines(t, out.String()) {
		got = append(got, r.level+" "+r.msg)
	}
	want := []string{"DEBUG trace", "WARN notice", "INFO shown"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("records = %v, want %v", got, want)
	}
}

func TestHertzCaller(t *testing.T) {
	out := log.CaptureForTest(t)
	log.WithHertz()
	caller := here(1)
	hlog.CtxInfof(t.Context(), "via hertz")

	records := parseLines(t, out.String())
	if len(records) != 1 || records[0].caller != caller {
		t.Fatalf("records = %+v, want caller %s", records, caller)
	}
}

func TestAppendLogKvLeavesParentUntouched(t *testing.T) {
	parent := log.AppendLogKv(t.Context(), "conn", "1")
	child := log.AppendLogKv(parent, "event", "push")
	sibling := log.AppendLogKv(parent, "event", "presence")

	if got := log.GetAllCustomFields(parent); len(got) != 1 {
		t.Fatalf("parent fields = %v", got)
	}
	if log.GetAllCustomFields(child)["event"] != "push" || log.GetAllCustomFields(sibling)["event"] != "presence" {
		t.Fatalf("child = %v, sibling = %v", log.GetAllCustomFields(child), log.GetAllCustomFields(sibling))
	}
	if got := log.AppendLogExtras(parent, nil); got != parent {
		t.Fatal("empty extras should return the parent context")
	}
}

func TestAppendLogKvConcurrentChildren(t *testing.T) {
	out := log.CaptureForTest(t)
	parent := log.AppendLogKv(t.Context(), "conn", "1")
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			log.CtxInfo(log.AppendLogKv(parent, "i", fmt.Sprint(i)), "child")
		})
	}
	wg.Wait()
	if n := len(parseLines(t, out.String())); n != 8 {
		t.Fatalf("got %d records, want 8", n)
	}
}

type sinkRecord struct {
	ctx   context.Context
	level log.Level
	msg   string
}

type captureSink struct{ records []sinkRecord }

func (s *captureSink) Emit(ctx context.Context, level log.Level, msg string) {
	s.records = append(s.records, sinkRecord{ctx, level, msg})
}

type ctxKey struct{}

func TestSinkReceivesFilteredRecords(t *testing.T) {
	log.CaptureForTest(t)
	sink := &captureSink{}
	log.AddSink(sink)
	log.SetLevel(log.LevelInfo)
	ctx := context.WithValue(t.Context(), ctxKey{}, "v")

	log.CtxDebug(ctx, "hidden")
	log.CtxWarn(ctx, "disk %d%%", 91)
	log.Info("plain")

	if len(sink.records) != 2 {
		t.Fatalf("sink records = %+v", sink.records)
	}
	warn := sink.records[0]
	if warn.level != log.LevelWarn || warn.msg != "disk 91%" || warn.ctx.Value(ctxKey{}) != "v" {
		t.Fatalf("warn record = %+v", warn)
	}
	if plain := sink.records[1]; plain.level != log.LevelInfo || plain.ctx == nil {
		t.Fatalf("plain record = %+v", plain)
	}
}

func TestSinkSkipsReplacedLogger(t *testing.T) {
	log.CaptureForTest(t)
	sink := &captureSink{}
	log.AddSink(sink)
	log.SetLogger(log.GetLogger())
	log.Info("still built-in")
	if len(sink.records) != 1 {
		t.Fatalf("sink records = %+v", sink.records)
	}
}

func TestErrorMarksRecordingSpan(t *testing.T) {
	log.CaptureForTest(t)
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	ctx, span := tp.Tracer("test").Start(t.Context(), "op")
	log.CtxWarn(ctx, "warn only")
	log.CtxError(ctx, "boom %s", "x")
	span.End()

	ended := recorder.Ended()
	if len(ended) != 1 {
		t.Fatalf("ended spans = %d", len(ended))
	}
	if got := ended[0].Status(); got.Code != codes.Error {
		t.Fatalf("span status = %+v, want error", got)
	}
	if events := ended[0].Events(); len(events) != 1 || events[0].Name != "exception" {
		t.Fatalf("span events = %+v, want one exception", events)
	}
}
