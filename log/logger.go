package log

import (
	"context"
	"io"
	"os"

	"github.com/cloudwego/kitex/pkg/klog"
	kitexlogrus "github.com/kitex-contrib/obs-opentelemetry/logging/logrus"
	"github.com/natefinch/lumberjack"
)

var (
	logger        Logger
	defaultLogger klog.FullLogger
	logLevel      Level
)

// Set custom format
func init() {
	logger = newLogger()
	logger.Logger.Logger().SetFormatter(&Formatter{})
	logger.SetLevel(klog.LevelDebug)
	logLevel = LevelDebug

	logger.Logger.Logger().AddHook(&traceIdHook{})
	defaultLogger = logger
}

func SetLogger(fullLogger klog.FullLogger) {
	defaultLogger = fullLogger
}

func SetProdEnv() {
	logger.SetLevel(klog.LevelInfo)
	logLevel = LevelInfo
	logger.Logger.Logger().AddHook(metricHook{})
}

type Logger struct {
	*kitexlogrus.Logger
}

func newLogger() Logger {
	return Logger{
		kitexlogrus.NewLogger(),
	}
}

func GetLogger() Logger {
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

// SetLevel sets the level of logs below which logs will not be output.
// The default log level is LevelTrace.
// Note that this method is not concurrent-safe.
func SetLevel(level Level) {
	var lv klog.Level
	switch level {
	case LevelTrace:
		lv = klog.LevelTrace
	case LevelDebug:
		lv = klog.LevelDebug
	case LevelInfo:
		lv = klog.LevelInfo
	case LevelWarn:
		lv = klog.LevelWarn
	case LevelError:
		lv = klog.LevelError
	case LevelFatal:
		lv = klog.LevelFatal
	default:
		lv = klog.LevelWarn
	}
	defaultLogger.SetLevel(lv)
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

// SetOutput sets the output of default logger. By default, it is stderr.
func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

// Fatal calls the default logger's Fatalf method and then os.Exit(1).
func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

// Error calls the default logger's Errorf method.
func Error(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

// Warn calls the default logger's Warnf method.
func Warn(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

// Notice calls the default logger's Noticef method.
func Notice(format string, v ...interface{}) {
	defaultLogger.Noticef(format, v...)
}

// Info calls the default logger's Infof method.
func Info(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

// Debug calls the default logger's Debugf method.
func Debug(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

// Trace calls the default logger's Tracef method.
func Trace(format string, v ...interface{}) {
	defaultLogger.Tracef(format, v...)
}

// CtxFatal calls the default logger's CtxFatalf method and then os.Exit(1).
func CtxFatal(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.CtxFatalf(ctx, format, v...)
}

// CtxError calls the default logger's CtxErrorf method.
func CtxError(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.CtxErrorf(ctx, format, v...)
}

// CtxWarn calls the default logger's CtxWarnf method.
func CtxWarn(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.CtxWarnf(ctx, format, v...)
}

// CtxNotice calls the default logger's CtxNoticef method.
func CtxNotice(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.CtxNoticef(ctx, format, v...)
}

// CtxInfo calls the default logger's CtxInfof method.
func CtxInfo(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.CtxInfof(ctx, format, v...)
}

// CtxDebug calls the default logger's CtxDebugf method.
func CtxDebug(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.CtxDebugf(ctx, format, v...)
}

// CtxTrace calls the default logger's CtxTracef method.
func CtxTrace(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.CtxTracef(ctx, format, v...)
}

func GetLogLevel() Level {
	return logLevel
}
