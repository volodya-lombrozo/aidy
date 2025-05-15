package main

import (
	"bytes"
	"fmt"
	"io"
	"sort"
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

	heal(mockGit)

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

	clean()

	_, err = os.Stat(aidyDir)
	assert.True(t, os.IsNotExist(err), ".aidy directory should be removed")
}

func TestStart(t *testing.T) {
	brain := ai.NewMockAI()
	shell := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	gh := &github.MockGithub{}

	err := startIssue("42", brain, &git.MockGit{Shell: shell}, gh)

	assert.NoError(t, err, "Expected no error when starting issue")
	expected := []string{"git checkout -b 42-mock-branch-name"}
	assert.Equal(t, len(expected), len(shell.Commands), "number of commands should match")
	for i, expectedCommand := range expected {
		assert.Contains(t, shell.Commands[i], expectedCommand)
	}
}

func TestStartIssueNoNumber(t *testing.T) {
	brain := ai.NewMockAI()
	shell := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	gh := &github.MockGithub{}

	err := startIssue("", brain, &git.MockGit{Shell: shell}, gh)

	assert.Error(t, err, "expected error when no issue number is provided")
	assert.Contains(t, err.Error(), "error: no issue number provided")
}

func TestStartIssueInvalidNumber(t *testing.T) {
	brain := ai.NewMockAI()
	shell := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	gh := &github.MockGithub{}

	err := startIssue("invalid", brain, &git.MockGit{Shell: shell}, gh)

	assert.Error(t, err, "Expected error when an invalid issue number is provided")
	assert.Contains(t, err.Error(), "error: invalid issue number 'invalid'")
}

func TestStartIssueBranchNameError(t *testing.T) {
	brain := ai.NewFailedMockAI()
	shell := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	gh := &github.MockGithub{}

	err := startIssue("42", brain, &git.MockGit{Shell: shell}, gh)

	assert.Error(t, err, "expected error when generating branch name fails")
	assert.Contains(t, err.Error(), "error generating branch name")
	assert.Contains(t, err.Error(), "failed to suggest branch")
}

func TestStartIssueCheckoutError(t *testing.T) {
	brain := ai.NewMockAI()
	shell := &executor.MockExecutor{
		Output: "",
		Err:    fmt.Errorf("error checking out branch"),
	}
	gh := &github.MockGithub{}
	mockGit := &git.MockGit{Shell: shell}

	err := startIssue("42", brain, mockGit, gh)

	assert.Error(t, err, "expected error when checking out branch fails")
	assert.Contains(t, err.Error(), "error checking out branch")
}

func TestSquash(t *testing.T) {
	mockAI := ai.NewMockAI()
	mockExecutor := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	mockGit := &git.MockGit{Shell: mockExecutor}

	squash(mockGit, mockAI)

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
	mockAI := ai.NewMockAI()
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
	mockAI := ai.NewMockAI()
	mockExecutor := &executor.MockExecutor{
		Output: "",
		Err:    nil,
	}
	mockGit := &git.MockGit{Shell: mockExecutor}

	commit(mockGit, mockAI)

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
	mockAI := ai.NewMockAI()
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
		result := inumber(test.branchName)
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

func TestLatestTag(t *testing.T) {
	versions := []string{"v1.0.0", "v1.1.0", "2.0.0"}

	tag := latest(versions)

	assert.Equal(t, "v1.1.0", tag, "Expected latest tag to be 'v1.1.0'")
}

func TestKeys(t *testing.T) {
	input := map[string]string{
		"feat(#42)": "Add clean command to clear cache",
		"fix(#42)":  "Fix bug in clean command",
	}

	result := keys(input)

	expected := []string{"feat(#42)", "fix(#42)"}
	sort.Strings(result)
	assert.Equal(t, expected, result)
}

func TestUpver(t *testing.T) {
	tests := []struct {
		actual   string
		step     string
		expected string
	}{
		{"v1.0.0", "patch", "1.0.1"},
		{"v1.0.0", "minor", "1.1.0"},
		{"v1.0.0", "major", "2.0.0"},
	}

	for _, test := range tests {
		tag, err := upver(test.actual, test.step)
		require.NoError(t, err, "Should upgrade the version successfully")
		assert.Equal(t, test.expected, tag)
	}

}
func TestReleaseSuccess(t *testing.T) {
	mgit := &git.MockGit{}
	nobrain := ai.NewMockAI()
	out := output.NewMock()

	err := release("minor", mgit, nobrain, out)
	assert.NoError(t, err, "expected no error during release")
	expected := "git tag -a \"v2.1.0\" -m \""
	assert.Contains(t, out.Last(), expected, "expected release command to be generated")
}

func TestReleaseNoTags(t *testing.T) {
	mockGit := &git.MockGit{}
	mockAI := ai.NewMockAI()
	out := output.NewMock()

	err := release("", mockGit, mockAI, out)

	assert.EqualError(t, err, "failed to update version: 'unknown version step: '''", "expected error when no tags are present")
}

func TestReleaseTagFetchError(t *testing.T) {
	mgit := &git.MockGit{
		Shell: &executor.MockExecutor{
			Output: "",
			Err:    fmt.Errorf("error fetching tags"),
		},
	}
	nobrain := ai.NewMockAI()
	out := output.NewMock()

	err := release("patch", mgit, nobrain, out)

	assert.Error(t, err, "expected error when fetching tags fails")
	assert.Contains(t, err.Error(), "failed to get tags", "expected error message about fetching tags")
}

func TestReleaseNotesGenerationError(t *testing.T) {
	mgit := &git.MockGit{}
	nobrain := ai.NewFailedMockAI()
	out := output.NewMock()

	err := release("major", mgit, nobrain, out)

	assert.Error(t, err, "expected error when generating release notes fails")
	assert.Contains(t, err.Error(), "failed to generate release notes", "expected error message about release notes generation")
}
