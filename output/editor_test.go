package output

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/executor"
)

func TestEditor_Print_RunOption(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	editor.in = r
	_, err := io.WriteString(w, "r\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	editor.Print(command)

	assert.Len(t, shell.Commands, 1, "expected 1 command to be run")
	assert.Equal(t, "echo 'Hello, World!'", shell.Commands[0], "expected command to match")
}

func TestEditor_Print_PrintOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	editor.in = input_r
	editor.out = output_w
	_, err := io.WriteString(input_w, "p\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	editor.Print(command)

	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), command, "expected command in output")
	assert.Len(t, shell.Commands, 0, "expected no command to be run")
}

func TestEditor_Print_CancelOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	editor.in = input_r
	editor.out = output_w
	_, err := io.WriteString(input_w, "c\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	editor.Print(command)

	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), "canceled", "expected cancel message in output")
	assert.Len(t, shell.Commands, 0, "expected no command to be run")
}

func TestEditor_Print_EditOption(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	editor.in = r
	_, err := io.WriteString(w, "e\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")

	command := "echo 'Hello, World!'"
	editor.Print(command)

	assert.Len(t, shell.Commands, 2, "expected 2 commands to be run")
	assert.Equal(t, command, shell.Commands[1], "expected edited command to match")
}
