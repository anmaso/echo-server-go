package logger

var (
	defaultLogger = New(INFO)
)

// Global logger functions
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// SetLevel sets the logging level for the default logger
func SetLevel(level Level) {
	defaultLogger.level = level
}
