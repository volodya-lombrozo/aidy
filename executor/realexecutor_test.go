package executor

import (
	"os"
	"strings"
	"testing"
)

func TestRealExecutor_RunCommand(t *testing.T) {
	executor := NewRealExecutor()

	// Test the echo command
	output, err := executor.RunCommand("echo", "Hello, World!")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Trim the output to remove any trailing newline characters
	output = strings.TrimSpace(output)

	expectedOutput := "Hello, World!"
	if output != expectedOutput {
		t.Fatalf("Expected '%s', got '%s'", expectedOutput, output)
	}
}

func TestRealExecutor_RunCommandInDir(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "execdirtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Error removing temp directory: %v", err)
		}
	}()

	executor := NewRealExecutor()
	output, err := executor.RunCommandInDir(tempDir, "echo", "Hello, Directory!")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output = strings.TrimSpace(output)
	expectedOutput := "Hello, Directory!"
	if output != expectedOutput {
		t.Fatalf("Expected '%s', got '%s'", expectedOutput, output)
	}
}
