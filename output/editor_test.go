package output

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volodya-lombrozo/aidy/executor"
)

func TestEditor_Print(t *testing.T) {
	shell := executor.NewMock()
	editor := NewEditor(shell)
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	if _, err := io.WriteString(w, "r\n"); err != nil {
		t.Errorf("failed to write to pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("failed to close write pipe: %v", err)
	}

	command := "echo 'Hello, World!'"

	editor.Print(command)

	assert.Len(t, shell.Commands, 1, "expected 1 command to be run")
	assert.Equal(t, "echo 'Hello, World!'", shell.Commands[0], "expected command to match")
}

func TestEditor_Print_DefaultOption(t *testing.T) {
	shell := executor.NewMock()
	editor := NewEditor(shell)
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	if _, err := io.WriteString(w, "r\n"); err != nil {
		t.Errorf("failed to write to pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("failed to close write pipe: %v", err)
	}
	command := "echo 'Hello, World!'"

	editor.Print(command)

	assert.Len(t, shell.Commands, 1, "expected 1 command to be run")
	assert.Equal(t, "echo 'Hello, World!'", shell.Commands[0], "expected command to match")
}

func TestEditor_Print_PrintOption(t *testing.T) {
	shell := executor.NewMock()
	editor := NewEditor(shell)
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	if _, err := io.WriteString(w, "p\n"); err != nil {
		t.Errorf("failed to write to pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("failed to close write pipe: %v", err)
	}
	command := "echo 'Hello, World!'"

	editor.Print(command)

	assert.Len(t, shell.Commands, 0, "expected no command to be run")
}

func TestEditor_Print_CancelOption(t *testing.T) {
	shell := executor.NewMock()
	editor := NewEditor(shell)
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	if _, err := io.WriteString(w, "c\n"); err != nil {
		t.Errorf("failed to write to pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("failed to close write pipe: %v", err)
	}
	command := "echo 'Hello, World!'"

	editor.Print(command)

	assert.Len(t, shell.Commands, 0, "expected no command to be run")
}

func TestEditor_Print_EditOption(t *testing.T) {
	shell := executor.NewMock()
	editor := NewEditor(shell)
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	if _, err := io.WriteString(w, "e\n"); err != nil {
		t.Errorf("failed to write to pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("failed to close write pipe: %v", err)
	}

	command := "echo 'Hello, World!'"
	editor.Print(command)

	assert.Len(t, shell.Commands, 2, "expected 2 commands to be run")
	assert.Equal(t, command, shell.Commands[1], "expected edited command to match")
}
