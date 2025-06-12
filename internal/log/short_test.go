package log

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShort_Info(t *testing.T) {
	mock := NewMock()
	short := NewShort(mock)

	long := strings.Repeat("a", 150)
	short.Info(long)

	assert.Len(t, mock.Messages, 1, "Expected one message to be logged")
	assert.True(t, len(mock.Messages[0]) <= 120+len("mcok info: ")+len("..."), "Expected message to be truncated to 120 characters")
	assert.Contains(t, mock.Messages[0], "mock info: ", "Expected info message to be logged")
}

func TestShort_Debug(t *testing.T) {
	mock := NewMock()
	short := NewShort(mock)

	long := strings.Repeat("b", 150)
	short.Debug(long)

	assert.Len(t, mock.Messages, 1, "Expected one message to be logged")
	assert.True(t, len(mock.Messages[0]) <= 120+len("mcok debug: ")+len("..."), "Expected message to be truncated to 120 characters")
	assert.Contains(t, mock.Messages[0], "mock dubug: ", "Expected debug message to be logged")
}

func TestShort_Warn(t *testing.T) {
	mockLogger := NewMock()
	short := NewShort(mockLogger)

	long := strings.Repeat("c", 150)
	short.Warn(long)

	assert.Len(t, mockLogger.Messages, 1, "Expected one message to be logged")
	assert.True(t, len(mockLogger.Messages[0]) <= 120+len("mcok warn: ")+len("..."), "Expected message to be truncated to 120 characters")
	assert.Contains(t, mockLogger.Messages[0], "mock warn: ", "Expected warning message to be logged")
}
