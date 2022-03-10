package decision

import (
	"github.com/sirupsen/logrus"
)

var logger Logger = &DefaultLogger{
	Entry: logrus.New().WithField("component", "common"),
}

// Level type is an iota that represents log level
type Level uint32

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

// Logger interface represent an object that can log a string and specify the minimal level for the output
type Logger interface {
	Logf(level Level, format string, args ...interface{})
	SetLevel(level Level)
}

// SetLogger sets the current logger
func SetLogger(l Logger) {
	logger = l
}

// SetLevel sets the level of the current logger
func SetLevel(level Level) {
	logger.SetLevel(level)
}

var lvlLogrusMap = map[Level]logrus.Level{
	PanicLevel: logrus.PanicLevel,
	FatalLevel: logrus.FatalLevel,
	ErrorLevel: logrus.ErrorLevel,
	WarnLevel:  logrus.WarnLevel,
	InfoLevel:  logrus.InfoLevel,
	DebugLevel: logrus.DebugLevel,
	TraceLevel: logrus.TraceLevel,
}

// DefaultLogger is a logrus based logger
type DefaultLogger struct {
	*logrus.Entry
}

// Logf logs a formatted string and specify its level
func (l *DefaultLogger) Logf(level Level, format string, args ...interface{}) {
	l.Entry.Logf(lvlLogrusMap[level], format, args...)
}

// SetLevel specifies the minimum log level to log to the logging output
func (l *DefaultLogger) SetLevel(level Level) {
	l.Logger.SetLevel(lvlLogrusMap[level])
}
