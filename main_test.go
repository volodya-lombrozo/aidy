package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/volodya-lombrozo/aidy/ai"
	"github.com/volodya-lombrozo/aidy/executor"
	"github.com/volodya-lombrozo/aidy/git"
	"github.com/volodya-lombrozo/aidy/github"
)

func TestHeal(t *testing.T) {
	mockGit := &git.MockGit{}
	mockExecutor := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}

	heal(mockGit, mockExecutor)

	expectedCommands := []string{
		"git commit --amend -m feat(#41): current commit message",
	}
	for i, expectedCommand := range expectedCommands {
		if !strings.Contains(mockExecutor.Commands[i], expectedCommand) {
			t.Errorf("Expected command '%s', got '%s'", expectedCommand, mockExecutor.Commands[i])
		}
	}
}

func TestSquash(t *testing.T) {
	mockGit := &git.MockGit{}
	mockAI := &ai.MockAI{}
	mockExecutor := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}

	squash(mockGit, mockExecutor, mockAI)

	expectedCommands := []string{
		"git reset --soft main",
        "git add --all",
		"git commit --amend -m feat(#41): current commit message",
	}

	for i, expectedCommand := range expectedCommands {
		if !strings.Contains(mockExecutor.Commands[i], expectedCommand) {
			t.Errorf("Expected command '%s', got '%s'", expectedCommand, mockExecutor.Commands[i])
		}
	}
}

func TestPullRequest(t *testing.T) {
	mockGit := &git.MockGit{}
	mockAI := &ai.MockAI{}
	mockGithub := &github.MockGithub{}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	pull_request(mockGit, mockAI, mockGithub)

	if err := w.Close(); err != nil {
		t.Fatalf("Error closing pipe writer: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Error copying data: %v", err)
	}
	output := buf.String()

	expected := "\ngh pr create --title \"Mock Title for 41_working_branch\" --body \"Mock Body for 41_working_branch\""
	if strings.TrimSpace(output) != strings.TrimSpace(expected) {
		t.Errorf("Unexpected output:\n%s", output)
	}
}

func TestCommit(t *testing.T) {
	mockGit := &git.MockGit{}
	mockAI := &ai.MockAI{}
	mockExecutor := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}

	commit(mockGit, mockExecutor, false, mockAI)

	expectedCommands := []string{
        "git add --all",
		"git commit --amend -m feat(#41): current commit message",
	}
	for i, expectedCommand := range expectedCommands {
		if !strings.Contains(mockExecutor.Commands[i], expectedCommand) {
			t.Errorf("Expected command '%s', got '%s'", expectedCommand, mockExecutor.Commands[i])
		}
	}
}

func TestHandleIssue(t *testing.T) {
	mockAI := &ai.MockAI{}
	gh := &github.MockGithub{}
	userInput := "test input"
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	issue(userInput, mockAI, gh)
	if err := w.Close(); err != nil {
		t.Fatalf("Error closing pipe writer: %v", err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Error copying data: %v", err)
	}
	output := buf.String()
	expected := "\ngh issue create --title \"Mock Issue Title for test input\" --body \"Mock Issue Body for test input\" --label \"bug,documentation,question\""
	if strings.TrimSpace(output) != strings.TrimSpace(expected) {
		t.Errorf("Unexpected output:\n%s", output)
	}
}

func TestHandleHelp(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	help()
	if err := w.Close(); err != nil {
		t.Fatalf("Error closing pipe writer: %v", err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Error copying data: %v", err)
	}
	output := buf.String()
	expected := `Usage:
  aidy pr   - Generate a pull request using AI-generated title and body.
  aidy help - Show this help message.`
	if strings.TrimSpace(output) != strings.TrimSpace(expected) {
		t.Errorf("Unexpected output:\n%s", output)
	}
}

func TestExtractIssueNumber(t *testing.T) {
	tests := []struct {
		branchName string
		expected   string
	}{
		{"123_feature", "123"},
		{"456_bugfix", "456"},
		{"789", "789"},
		{"no_issue_number", "no"},
		{"", "unknown"},
	}

	for _, test := range tests {
		result := extractIssueNumber(test.branchName)
		if result != test.expected {
			t.Errorf("For branch name '%s', expected '%s', got '%s'", test.branchName, test.expected, result)
		}
	}
}

func TestEscapeBackticks(t *testing.T) {
	input := "This is a `test` string with `backticks`."
	expected := "This is a \\`test\\` string with \\`backticks\\`."
	result := escapeBackticks(input)

	if result != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, result)
	}
}
