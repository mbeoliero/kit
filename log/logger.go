package log

import (
	"context"
	"io"
	"os"

	"github.com/cloudwego/kitex/pkg/klog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	builtin                       = newTextLogger(os.Stdout)
	logger                        = &Logger{FullLogger: builtin}
	defaultLogger klog.FullLogger = logger
	logLevel      Level
)

// Logger wraps the logger that the package-level functions write to.
type Logger struct {
	klog.FullLogger
}

func init() {
	SetLevel(LevelDebug)
}

// SetLogger replaces the logger behind the package-level functions, typically to
// capture output in tests. Sinks only observe the built-in logger.
func SetLogger(fullLogger klog.FullLogger) {
	defaultLogger = fullLogger
}

// SetProdEnv applies production defaults: info level on the built-in logger.
func SetProdEnv() {
	logger.SetLevel(klog.LevelInfo)
	logLevel = LevelInfo
}

func GetLogger() *Logger {
	return logger
}

// Level defines the priority of a log message.
// When a logger is configured with a level, any log message with a lower
// log level (smaller by integer comparison) will not be output.
type Level int

// The levels of logs.
const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func (l Level) toKlog() klog.Level {
	switch l {
	case LevelTrace:
		return klog.LevelTrace
	case LevelDebug:
		return klog.LevelDebug
	case LevelInfo:
		return klog.LevelInfo
	case LevelWarn:
		return klog.LevelWarn
	case LevelError:
		return klog.LevelError
	case LevelFatal:
		return klog.LevelFatal
	default:
		return klog.LevelWarn
	}
}

// fromKlogLevel maps a threshold; Notice has no own level and filters like Warn.
func fromKlogLevel(level klog.Level) Level {
	switch level {
	case klog.LevelTrace:
		return LevelTrace
	case klog.LevelDebug:
		return LevelDebug
	case klog.LevelInfo:
		return LevelInfo
	case klog.LevelNotice, klog.LevelWarn:
		return LevelWarn
	case klog.LevelError:
		return LevelError
	case klog.LevelFatal:
		return LevelFatal
	default:
		return LevelWarn
	}
}

// recordLevel maps a record's level. Trace records are written and filtered as Debug,
// matching the previous zerolog backend, so a debug threshold still shows them.
func recordLevel(level klog.Level) Level {
	if level == klog.LevelTrace {
		return LevelDebug
	}
	return fromKlogLevel(level)
}

// SetLevel sets the level of logs below which logs will not be output.
// The default log level is LevelDebug.
// Note that this method is not concurrent-safe.
func SetLevel(level Level) {
	defaultLogger.SetLevel(level.toKlog())
	logLevel = level
}

// SetLogFile sets log output to file and stdout.
// Use lumberjack to rolling file.
func SetLogFile(fileName string, ops ...LogfileOption) {
	// roller with default params
	rollingWriter := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    256,  // Single file max capacity, MB
		MaxBackups: 20,   // Maximum number of expired files to keep
		MaxAge:     10,   // Maximum days to keep expired files
		Compress:   true, // Whether rolling logs need to be compressed, use gzip to compress
	}

	for _, op := range ops {
		op.apply(rollingWriter)
	}

	mw := io.MultiWriter(rollingWriter, os.Stdout)
	defaultLogger.SetOutput(mw)
}

// SetOutput sets the output of default logger. By default, it is stdout.
func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

// Fatal calls the default logger's Fatalf method and then os.Exit(1).
func Fatal(format string, v ...any) {
	defaultLogger.Fatalf(format, v...)
}

// Error calls the default logger's Errorf method.
func Error(format string, v ...any) {
	defaultLogger.Errorf(format, v...)
}

// Warn calls the default logger's Warnf method.
func Warn(format string, v ...any) {
	defaultLogger.Warnf(format, v...)
}

// Notice calls the default logger's Noticef method.
func Notice(format string, v ...any) {
	defaultLogger.Noticef(format, v...)
}

// Info calls the default logger's Infof method.
func Info(format string, v ...any) {
	defaultLogger.Infof(format, v...)
}

// Debug calls the default logger's Debugf method.
func Debug(format string, v ...any) {
	defaultLogger.Debugf(format, v...)
}

// Trace calls the default logger's Tracef method.
func Trace(format string, v ...any) {
	defaultLogger.Tracef(format, v...)
}

// CtxFatal calls the default logger's CtxFatalf method and then os.Exit(1).
func CtxFatal(ctx context.Context, format string, v ...any) {
	defaultLogger.CtxFatalf(ctx, format, v...)
}

// CtxError calls the default logger's CtxErrorf method.
func CtxError(ctx context.Context, format string, v ...any) {
	defaultLogger.CtxErrorf(ctx, format, v...)
}

// CtxWarn calls the default logger's CtxWarnf method.
func CtxWarn(ctx context.Context, format string, v ...any) {
	defaultLogger.CtxWarnf(ctx, format, v...)
}

// CtxNotice calls the default logger's CtxNoticef method.
func CtxNotice(ctx context.Context, format string, v ...any) {
	defaultLogger.CtxNoticef(ctx, format, v...)
}

// CtxInfo calls the default logger's CtxInfof method.
func CtxInfo(ctx context.Context, format string, v ...any) {
	defaultLogger.CtxInfof(ctx, format, v...)
}

// CtxDebug calls the default logger's CtxDebugf method.
func CtxDebug(ctx context.Context, format string, v ...any) {
	defaultLogger.CtxDebugf(ctx, format, v...)
}

// CtxTrace calls the default logger's CtxTracef method.
func CtxTrace(ctx context.Context, format string, v ...any) {
	defaultLogger.CtxTracef(ctx, format, v...)
}

func GetLogLevel() Level {
	return logLevel
}
