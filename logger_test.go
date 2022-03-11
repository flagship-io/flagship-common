package decision

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetLogger(t *testing.T) {
	newLogger := &DefaultLogger{}
	SetLogger(newLogger)
	assert.Equal(t, newLogger, logger)
}

func TestSetLogLevel(t *testing.T) {
	l := &DefaultLogger{
		Entry: logrus.New().WithField("component", "common"),
	}
	l.SetLevel(InfoLevel)
	assert.Equal(t, logrus.InfoLevel, l.Logger.Level)
}

func TestLogf(t *testing.T) {
	l := &DefaultLogger{
		Entry: logrus.New().WithField("component", "common"),
	}

	var b bytes.Buffer
	mockWriter := bufio.NewWriter(&b)
	l.Logger.SetOutput(mockWriter)

	l.Logf(DebugLevel, "test %v", "value")
	mockWriter.Flush()
	ret := b.Bytes()
	assert.Equal(t, "", string(ret))

	b.Reset()
	l.SetLevel(DebugLevel)

	l.Logf(DebugLevel, "test %v", "value")
	mockWriter.Flush()
	ret = b.Bytes()
	assert.Contains(t, string(ret), "test value")
}
