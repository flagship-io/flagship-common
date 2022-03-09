package decision

import (
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
	l := &DefaultLogger{}
	err := l.SetLogLevel("wrong")
	assert.NotNil(t, err)

	err = l.SetLogLevel("info")
	assert.Nil(t, err)
	assert.Equal(t, logrus.InfoLevel, l.Level)
}
