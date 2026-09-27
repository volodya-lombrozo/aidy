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

func TestTextReviewer_Review_AcceptOption(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewTextReviewer(shell)
	reviewer.in = r
	_, err := io.WriteString(w, "a\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := reviewer.Review(Fixed(text))

	require.NoError(t, err, "Review should not return an error")
	assert.Equal(t, text, result, "expected the original text to be returned unchanged")
	assert.Len(t, shell.Commands, 0, "expected no external editor to be run")
}

func TestTextReviewer_Review_CancelOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewTextReviewer(shell)
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "c\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := reviewer.Review(Fixed(text))

	assert.ErrorIs(t, err, ErrCanceled, "expected a canceled error")
	assert.Empty(t, result, "expected no text to be returned when canceled")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), "canceled", "expected cancel message in output")
}

func TestTextReviewer_Review_PrintOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewTextReviewer(shell)
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "p\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := reviewer.Review(Fixed(text))

	assert.ErrorIs(t, err, ErrCanceled, "expected a canceled error")
	assert.Empty(t, result, "expected no text to be returned after printing")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), text, "expected text in output")
}

func TestTextReviewer_Review_EditOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewTextReviewer(shell)
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "e\na\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := reviewer.Review(Fixed(text))

	require.NoError(t, err, "Review should not return an error")
	assert.Equal(t, text, result, "expected the unmodified temp file content to be returned")
	assert.Len(t, shell.Commands, 1, "expected the external editor to be run once")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), "updated", "expected updated text label in output")
}

func TestTextReviewer_Review_EditOption_FailsWithError(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("simulated error")
	reviewer := NewTextReviewer(shell)
	reviewer.in = r
	_, err := io.WriteString(w, "e\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	_, err = reviewer.Review(Fixed(text))

	assert.Error(t, err, "expected an error when the external editor fails")
	assert.Contains(t, err.Error(), "simulated error", "expected error message to match")
	assert.Contains(t, err.Error(), "failed to edit text", "expected error to mention text editing failure")
}

func TestTextReviewer_Review_InvalidOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	err_r, err_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewTextReviewer(shell)
	reviewer.in = input_r
	reviewer.err = err_w
	_, err := io.WriteString(input_w, "x\na\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := reviewer.Review(Fixed(text))

	require.NoError(t, err, "Review should not return an error once a valid option is given")
	assert.Equal(t, text, result, "expected the original text to be returned unchanged")
	err = err_w.Close()
	require.NoError(t, err, "failed to close error pipe")
	output, err := io.ReadAll(err_r)
	require.NoError(t, err, "failed to read from error output")
	assert.Contains(t, string(output), "please type a, e, g, c, or p", "expected a hint about valid options")
}

func TestTextReviewer_Review_GenerateOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	reviewer := NewTextReviewer(executor.NewMock())
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "g\na\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	generated := 0
	notes := func() (string, error) {
		generated++
		return fmt.Sprintf("release notes version-%d", generated), nil
	}

	result, err := reviewer.Review(notes)

	require.NoError(t, err, "Review should not return an error")
	assert.Equal(t, 2, generated, "expected a second version to be generated on demand")
	assert.Equal(t, "release notes version-2", result, "expected the freshly generated text to be accepted")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), "[g]enerate", "expected the generate option to be offered")
}

func TestTextReviewer_Review_FailsWhenTextFails(t *testing.T) {
	reviewer := NewTextReviewer(executor.NewMock())
	notes := func() (string, error) {
		return "", fmt.Errorf("simulated error")
	}

	_, err := reviewer.Review(notes)

	require.Error(t, err, "expected an error when the generator fails")
	assert.Contains(t, err.Error(), "simulated error", "expected the generator error to be reported")
}

func TestTextReviewer_Review_GenerateOption_KeepsFixedText(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	reviewer := NewTextReviewer(executor.NewMock())
	reviewer.in = input_r
	_, err := io.WriteString(input_w, "g\na\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	text := "release notes text"

	result, err := reviewer.Review(Fixed(text))

	require.NoError(t, err, "Review should not return an error")
	assert.Equal(t, text, result, "expected a fixed text to stay the same when generated again")
}

func TestTextReviewer_Review_ReadError(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	err_r, err_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewTextReviewer(shell)
	reviewer.in = input_r
	reviewer.err = err_w
	require.NoError(t, input_w.Close(), "failed to close write pipe")
	text := "release notes text"

	_, err := reviewer.Review(Fixed(text))

	assert.Error(t, err, "expected an error when reading input fails")
	require.NoError(t, err_w.Close(), "failed to close error pipe")
	output, rerr := io.ReadAll(err_r)
	require.NoError(t, rerr, "failed to read from error output")
	assert.Contains(t, string(output), "Error reading input", "expected an input-reading error message")
}
