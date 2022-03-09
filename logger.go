package decision

import "github.com/sirupsen/logrus"

var logger Logger = &DefaultLogger{}

type Logger interface {
	Debug(args ...interface{})
	Trace(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	SetLogLevel(lvl string) error
}

func SetLogger(l Logger) {
	logger = l
}

type DefaultLogger struct {
	logrus.Logger
}

func (l *DefaultLogger) SetLogLevel(lvl string) error {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		return err
	}
	l.Level = level
	return err
}
