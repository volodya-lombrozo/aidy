package executor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockExecutor_RunCommand(t *testing.T) {
	mock := NewMock()
	mock.Output = "mock output"

	output, err := mock.RunCommand("any-command")

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, "mock output", output, "Expected 'mock output'")
}
func TestMockExecutor_RunInteractively(t *testing.T) {
	mock := NewMock()
	mock.Output = "mock output"

	output, err := mock.RunInteractively("any-command", "arg1", "arg2")

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, "mock output", output, "Expected 'mock output'")
	assert.Len(t, mock.Commands, 1, "Expected one command")
	assert.Equal(t, "any-command arg1 arg2", mock.Commands[0], "Expected command 'any-command arg1 arg2'")
}

func TestMockExecutor_RunCommandInDir(t *testing.T) {
	mock := NewMock()
	mock.Output = "mock output"

	output, err := mock.RunCommandInDir("/some/dir", "any-command", "arg1", "arg2")

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, "mock output", output, "Expected 'mock output'")
	assert.Len(t, mock.Commands, 1, "Expected one command")
	assert.Equal(t, "cd /some/dir && any-command arg1 arg2", mock.Commands[0], "Expected command 'cd /some/dir && any-command arg1 arg2'")
}

func TestMockExecutor_RunCommandWithError(t *testing.T) {
	mock := NewMock()
	mock.Err = fmt.Errorf("mock error")

	_, err := mock.RunCommand("any-command")

	assert.Error(t, err, "Expected error")
	assert.EqualError(t, err, "mock error", "Expected 'mock error'")
}
