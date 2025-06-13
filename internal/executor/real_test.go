package executor

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/log"
)

func TestRealExecutor_RunCommand(t *testing.T) {
	executor := NewReal()

	output, err := executor.RunCommand("echo", "Hello, World!")

	require.NoError(t, err, "Expected no error when running command")
	assert.Equal(t, "Hello, World!\n", output, "Expected output to match")
}

func TestRealExecutor_RunCommandInDir(t *testing.T) {
	tmp, err := os.MkdirTemp("", "execdirtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() { require.NoError(t, os.RemoveAll(tmp), "Failed to remove temp dir") }()
	executor := NewReal()

	output, err := executor.RunCommandInDir(tmp, "echo", "Hello, Directory!")

	require.NoError(t, err, "Expected no error when running command in directory")
	assert.Equal(t, "Hello, Directory!", strings.TrimSpace(output), "Expected output to match")
}

func TestRealExecutor_RunInteractively(t *testing.T) {
	executor := &RealExecutor{log: log.Get()}
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe for stdout")
	executor.out = w

	output, err := executor.RunInteractively("echo", "Hello, Interactive World!")

	require.NoError(t, w.Close(), "Expected no error when closing write pipe")
	require.NoError(t, err, "Expected no error when running interactive command")
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "Expected no error when running interactive command")
	assert.Equal(t, "Hello, Interactive World!\n", buf.String(), "Expected output to match")
	assert.Equal(t, "unimplemented", output, "Expected output to be 'unimplemented' for interactive commands")
}
