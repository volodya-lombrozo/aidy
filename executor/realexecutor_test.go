package executor

import (
	"strings"
	"testing"
)

func TestRealExecutor_RunCommand(t *testing.T) {
	executor := &RealExecutor{}

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
