package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunTaskAsync(t *testing.T) {
	functionHasRun := false
	c := RunTaskAsync(func() {
		functionHasRun = true
	})
	r := <-c
	assert.True(t, r)
	assert.True(t, functionHasRun)
}
