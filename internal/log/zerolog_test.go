package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZerolog_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZerolog(&buf)

	logger.Info("This is an info message")

	assert.Contains(t, buf.String(), "This is an info message", "Expected info message to be logged")
}

func TestZerolog_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZerolog(&buf)

	logger.Debug("This is a debug message")

	assert.Contains(t, buf.String(), "This is a debug message", "Expected debug message to be logged")
}

func TestZerolog_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZerolog(&buf)

	logger.Warn("This is a warning message")

	assert.Contains(t, buf.String(), "This is a warning message", "Expected warning message to be logged")
}

func TestZerolog_Info_Parametrised(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZerolog(&buf)
	logger.Info("This is an info message with param: %d", 42)
	assert.Contains(t, buf.String(), "This is an info message with param: 42", "Expected info message with parameter to be logged")
}
