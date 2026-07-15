package output

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/executor"
)

func TestTextEditor_Edit_AcceptOption(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewTextEditor(shell)
	editor.in = r
	_, err := io.WriteString(w, "a\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := editor.Edit(text)

	require.NoError(t, err, "Edit should not return an error")
	assert.Equal(t, text, result, "expected the original text to be returned unchanged")
	assert.Len(t, shell.Commands, 0, "expected no external editor to be run")
}

func TestTextEditor_Edit_CancelOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewTextEditor(shell)
	editor.in = input_r
	editor.out = output_w
	_, err := io.WriteString(input_w, "c\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := editor.Edit(text)

	assert.ErrorIs(t, err, ErrCanceled, "expected a canceled error")
	assert.Empty(t, result, "expected no text to be returned when canceled")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), "canceled", "expected cancel message in output")
}

func TestTextEditor_Edit_PrintOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewTextEditor(shell)
	editor.in = input_r
	editor.out = output_w
	_, err := io.WriteString(input_w, "p\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := editor.Edit(text)

	assert.ErrorIs(t, err, ErrCanceled, "expected a canceled error")
	assert.Empty(t, result, "expected no text to be returned after printing")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), text, "expected text in output")
}

func TestTextEditor_Edit_EditOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewTextEditor(shell)
	editor.in = input_r
	editor.out = output_w
	_, err := io.WriteString(input_w, "e\na\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := editor.Edit(text)

	require.NoError(t, err, "Edit should not return an error")
	assert.Equal(t, text, result, "expected the unmodified temp file content to be returned")
	assert.Len(t, shell.Commands, 1, "expected the external editor to be run once")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), "updated", "expected updated text label in output")
}

func TestTextEditor_Edit_EditOption_FailsWithError(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("simulated error")
	editor := NewTextEditor(shell)
	editor.in = r
	_, err := io.WriteString(w, "e\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	_, err = editor.Edit(text)

	assert.Error(t, err, "expected an error when the external editor fails")
	assert.Contains(t, err.Error(), "simulated error", "expected error message to match")
	assert.Contains(t, err.Error(), "failed to edit text", "expected error to mention text editing failure")
}

func TestTextEditor_Edit_InvalidOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	err_r, err_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewTextEditor(shell)
	editor.in = input_r
	editor.err = err_w
	_, err := io.WriteString(input_w, "x\na\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := editor.Edit(text)

	require.NoError(t, err, "Edit should not return an error once a valid option is given")
	assert.Equal(t, text, result, "expected the original text to be returned unchanged")
	err = err_w.Close()
	require.NoError(t, err, "failed to close error pipe")
	output, err := io.ReadAll(err_r)
	require.NoError(t, err, "failed to read from error output")
	assert.Contains(t, string(output), "please type a, e, c, or p", "expected a hint about valid options")
}

func TestTextEditor_Edit_ReadError(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	err_r, err_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewTextEditor(shell)
	editor.in = input_r
	editor.err = err_w
	require.NoError(t, input_w.Close(), "failed to close write pipe")
	text := "release notes text"

	_, err := editor.Edit(text)

	assert.Error(t, err, "expected an error when reading input fails")
	require.NoError(t, err_w.Close(), "failed to close error pipe")
	output, rerr := io.ReadAll(err_r)
	require.NoError(t, rerr, "failed to read from error output")
	assert.Contains(t, string(output), "Error reading input", "expected an input-reading error message")
}
