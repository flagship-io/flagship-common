package decision

import (
	"github.com/sirupsen/logrus"
)

var logger Logger = &DefaultLogger{
	Logger: logrus.New(),
}

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

type Logger interface {
	Logf(level Level, format string, args ...interface{})
	SetLevel(level Level)
}

func SetLogger(l Logger) {
	logger = l
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

type DefaultLogger struct {
	*logrus.Logger
}

func (l *DefaultLogger) Logf(level Level, format string, args ...interface{}) {
	l.Logger.Logf(lvlLogrusMap[level], format, args...)
}

func (l *DefaultLogger) SetLevel(level Level) {
	l.Logger.SetLevel(lvlLogrusMap[level])
}
