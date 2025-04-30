package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/ai"
	"github.com/volodya-lombrozo/aidy/cache"
	"github.com/volodya-lombrozo/aidy/executor"
	"github.com/volodya-lombrozo/aidy/git"
	"github.com/volodya-lombrozo/aidy/github"
	"github.com/volodya-lombrozo/aidy/output"
)

func TestHeal(t *testing.T) {
	mockExecutor := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	mockGit := &git.MockGit{Shell: mockExecutor}

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
	mockAI := &ai.MockAI{}
	mockExecutor := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	mockGit := &git.MockGit{Shell: mockExecutor}

	squash(mockGit, mockExecutor, mockAI)

	expectedCommands := []string{
		"git reset --soft refs/heads/main",
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
	out := output.NewMock()

	pull_request(mockGit, mockAI, mockGithub, cache.NewMockAidyCache(), out)

	output := out.Last()
	expected := "\ngh pr create --title \"Mock Title for 41\" --body \"Mock Body for 41\" --repo mock/remote"
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(output))
}

func TestHealQoutes(t *testing.T) {
	message := healQuotes("\"with \" qoutes\"")
	assert.Equal(t, "\"with \" qoutes\"", message)

	message = healQuotes("'with ' qoutes'")
	assert.Equal(t, "'with ' qoutes'", message)

	message = healQuotes("`with ` qoutes`")
	assert.Equal(t, "`with ` qoutes`", message)
}

func TestHealQuotesParametrized(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"`feat(#123):feature`", "feat(#123):feature"},
		{"`top level quotes should be removed`", "top level quotes should be removed"},
		{"`aidy pr` works incorrectly", "`aidy pr` works incorrectly"},
		{"789", "789"},
		{"`aidy ci` and `aidy issue` commands", "`aidy ci` and `aidy issue` commands"},
		{"`aidy ci` and `aidy issue`", "`aidy ci` and `aidy issue`"},
		{"", ""},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, healQuotes(test.input))
	}
}

func TestCommit(t *testing.T) {
	mockAI := &ai.MockAI{}
	mockExecutor := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	mockGit := &git.MockGit{Shell: mockExecutor}

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
	out := output.NewMock()

	issue(userInput, mockAI, gh, cache.NewMockAidyCache(), out)

	output := out.Last()
	expected := "\ngh issue create --title \"Mock Issue Title for test input\" --body \"Mock Issue Body for test input\" --label \"bug,documentation,question\" --repo mock/remote"
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(output))
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
		{"no_issue_number", "no_issue_number"},
		{"", "unknown"},
		{"feature/1234-user-authentication", "1234"},
		{"bugfix/987-missing-footer", "987"},
		{"hotfix/456-null-pointer-crash", "456"},
		{"chore/321-cleanup-logs", "321"},
		{"task/202-improve-error-messages", "202"},
		{"ui/1123-fix-modal-animation", "1123"},
		{"api/789-add-endpoint-for-login", "789"},
		{"docs/345-update-readme", "345"},
		{"db/234-add-index-to-users-table", "234"},
		{"infra/1001-refactor-docker-setup", "1001"},
		{"release/v1.2.3", "1"},
		{"release/2.0.0-beta", "2"},
		{"milestone/1.0-phase1", "1"},
		{"v2.0/feature/001-add-auth", "001"},
		{"v3/cleanup/009-legacy-scripts", "009"},
		{"experiment/142-new-cache-layer", "142"},
		{"test/555-sentry-integration", "555"},
		{"spike/777-redesign-login-flow", "777"},
		{"quickfix-4202-crash-on-load", "4202"},
		{"ticket_900-improve-ci-speed", "900"},
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
