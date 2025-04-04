package executor

import (
	"fmt"
	"testing"
)

func TestMockExecutor_RunCommand(t *testing.T) {
	mock := &MockExecutor{
		Output: "mock output",
		Err:    nil,
	}

	output, err := mock.RunCommand("any-command")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output != "mock output" {
		t.Fatalf("Expected 'mock output', got '%s'", output)
	}
}

func TestMockExecutor_RunCommandWithError(t *testing.T) {
	mock := &MockExecutor{
		Output: "",
		Err:    fmt.Errorf("mock error"),
	}

	_, err := mock.RunCommand("any-command")
	if err == nil {
		t.Fatal("Expected error, got none")
	}
	if err.Error() != "mock error" {
		t.Fatalf("Expected 'mock error', got '%v'", err)
	}
}
