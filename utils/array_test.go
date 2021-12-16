package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInStringArray(t *testing.T) {
	assert.Equal(t, true, IsInStringArray("test", []string{"test", "test2"}))
	assert.Equal(t, false, IsInStringArray("test", []string{"test1", "test2"}))
	assert.Equal(t, false, IsInStringArray("test", []string{}))
	assert.Equal(t, false, IsInStringArray("", []string{"test1", "test2"}))
}
