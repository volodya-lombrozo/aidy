package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/ai"
	"github.com/volodya-lombrozo/aidy/cache"
	"github.com/volodya-lombrozo/aidy/executor"
	"github.com/volodya-lombrozo/aidy/git"
	"github.com/volodya-lombrozo/aidy/github"
	"os"
	"path/filepath"
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

func TestCleanCache(t *testing.T) {
	tempDir := t.TempDir()
	aidyDir := filepath.Join(tempDir, ".aidy")
	err := os.Mkdir(aidyDir, 0755)
	require.NoError(t, err, "Failed to create .aidy directory")

	originalDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change working directory")

	cleanCache()

	_, err = os.Stat(aidyDir)
	assert.True(t, os.IsNotExist(err), ".aidy directory should be removed")
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

	pull_request(mockGit, mockAI, mockGithub, cache.NewMockAidyCache())

	if err := w.Close(); err != nil {
		t.Fatalf("Error closing pipe writer: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Error copying data: %v", err)
	}
	output := buf.String()

	expected := "\ngh pr create --title \"Mock Title for 41_working_branch\" --body \"Mock Body for 41_working_branch\" --repo mock/remote"
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(output))
}

func TestHealQoutes(t *testing.T) {
	message := healQoutes("\"with \" qoutes\"")
	assert.Equal(t, "with \" qoutes", message)
	message = healQoutes("'with ' qoutes'")
	assert.Equal(t, "with ' qoutes", message)
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
	issue(userInput, mockAI, gh, cache.NewMockAidyCache())
	err := w.Close()
	require.NoError(t, err)
	os.Stdout = old
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()
	expected := "\ngh issue create --title \"Mock Issue Title for test input\" --body \"Mock Issue Body for test input\" --label \"bug,documentation,question\" --repo mock/remote"
	assert.Equal(t, strings.TrimSpace(output), strings.TrimSpace(expected))
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

func TestHealPRTitle(t *testing.T) {
	tests := []struct {
		actual   string
		expected string
	}{
		{"feat(#75): Add clean command to clear cache", "feat(#42): Add clean command to clear cache"},
		{"feat(#master): Add clean command to clear cache", "feat(#master): Add clean command to clear cache"},
		{"feat(#7523): Add clean command to clear cache", "feat(#42): Add clean command to clear cache"},
		{"fix(#7523): Add clean command to clear cache", "fix(#42): Add clean command to clear cache"},
		{"test(#7523): Add clean command to clear cache", "test(#42): Add clean command to clear cache"},
	}
	for _, test := range tests {
		result := healPRTitle(test.actual, "42")
		assert.Equal(t, test.expected, result)
	}
}

func TestHealPRBody(t *testing.T) {
	tests := []struct {
		actual   string
		expected string
	}{
		{"Add clean command to clear cache\nCloses: #75", "Add clean command to clear cache\nCloses #42"},
		{"add clean command to clear cache\ncloses: #75", "add clean command to clear cache\ncloses #42"},
		{"add clean command to clear cache\ncloses #75", "add clean command to clear cache\ncloses #42"},
		{"Add clean command to clear cache", "Add clean command to clear cache"},
		{"Add clean command to clear cache\nCloses: #master", "Add clean command to clear cache\nCloses: #master"},
		{"Add clean command to clear cache\nRelated to #7523", "Add clean command to clear cache\nRelated to #42"},
		{"add clean command to clear cache\nrelated to #7523", "add clean command to clear cache\nrelated to #42"},
		{"add clean command to clear cache\nrelated to: #7523", "add clean command to clear cache\nrelated to #42"},
		{"add clean command to clear cache\ntest #7523", "add clean command to clear cache\ntest #7523"},
	}
	for _, test := range tests {
		result := healPRBody(test.actual, "42")
		assert.Equal(t, test.expected, result)
	}
}
