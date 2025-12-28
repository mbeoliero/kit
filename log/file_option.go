package log

import "github.com/natefinch/lumberjack"

// LogfileOption is the only way to config log file option.
type LogfileOption interface {
	apply(config *lumberjack.Logger)
}

type logFileOption func(config *lumberjack.Logger)

func (fo logFileOption) apply(config *lumberjack.Logger) {
	fo(config)
}

// WithMaxSize set log file's max size, MB
func WithMaxSize(size int) LogfileOption {
	return logFileOption(func(config *lumberjack.Logger) {
		config.MaxSize = size
	})
}

// WithMaxBackups set maximum number of expired files to keep
func WithMaxBackups(backups int) LogfileOption {
	return logFileOption(func(config *lumberjack.Logger) {
		config.MaxBackups = backups
	})
}

// WithMaxAge set maximum days to keep expired files
func WithMaxAge(age int) LogfileOption {
	return logFileOption(func(config *lumberjack.Logger) {
		config.MaxAge = age
	})
}
