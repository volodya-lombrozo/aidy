package aidy

import (
	"fmt"
	"strings"
	"testing"

	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/ai"
	"github.com/volodya-lombrozo/aidy/internal/cache"
	"github.com/volodya-lombrozo/aidy/internal/config"
	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/git"
	"github.com/volodya-lombrozo/aidy/internal/github"
	"github.com/volodya-lombrozo/aidy/internal/output"
)

func TestReal_SetTarget_IfOnlyOne(t *testing.T) {
	cache := cache.NewMockAidyCache()
	aidy := &real{git: git.NewMock(), config: config.NewMock(), cache: cache, printer: output.NewMock()}

	err := aidy.SetTarget()

	require.NoError(t, err, "Expected no error when setting target with only one remote")
	actual := cache.Remote()
	assert.Equal(t, "mock/remote", actual, "Expected remote to be set to 'mock/remote'")
}

func TestReal_SetTarget_GitError(t *testing.T) {
	cache := cache.NewMockAidyCache()
	cache.WithRemote("")
	aidy := &real{
		git:     git.NewMockWithError(fmt.Errorf("mock error")),
		config:  config.NewMock(),
		cache:   cache,
		printer: output.NewMock(),
	}

	err := aidy.SetTarget()

	assert.Error(t, err, "Expected error when setting target with git error")
}

func TestReal_SetTarget_PreservesDots(t *testing.T) {
	cache := cache.NewMockAidyCache()
	shell := executor.NewMock()
	shell.Output = "https://github.com/cqfn/kaicode.github.io.git"
	cache.WithRemote("")
	aidy := &real{
		git:     git.NewMockWithShell(shell),
		config:  config.NewMock(),
		cache:   cache,
		printer: output.NewMock(),
	}

	err := aidy.SetTarget()

	require.NoError(t, err, "Expected no error when setting target with dots in remote")
	assert.Equal(t, "cqfn/kaicode.github.io", cache.Remote(), "expected dots in remote to be preserved")
}

func TestReal_PrintsDiff_Successfully(t *testing.T) {
	printer := output.NewMock()
	git := git.NewMock()
	raidy := &real{git: git, config: config.NewMock(), printer: printer}

	err := raidy.Diff()

	captured := printer.Captured()
	require.NoError(t, err, "Expected no error when printing diff")
	expected := "Diff with the base branch:\nmock-diff\n"
	assert.Equal(t, expected, captured, "Expected diff to be printed")
}

func TestReal_PrintsDiff_Error(t *testing.T) {
	printer := output.NewMock()
	git := git.NewMockWithError(fmt.Errorf("mock error"))
	raidy := &real{git: git, config: config.NewMock(), printer: printer}

	err := raidy.Diff()

	require.Error(t, err, "Expected error when printing diff")
	assert.Equal(t, "failed to get diff: 'mock error'", err.Error(), "Expected error message to match")
}

func TestReal_PrintConfig_Successful(t *testing.T) {
	printer := output.NewMock()
	raidy := &real{config: config.NewMock(), printer: printer}

	err := raidy.PrintConfig()

	require.NoError(t, err, "Expected no error when printing configuration")
	res := printer.Captured()
	expected := strings.Join([]string{
		"Aidy Configuration:",
		"",
		"OpenAI API Key: mock-openai-key",
		"",
		"Deepseek API Key: mock-deepseek-key",
		"",
		"GitHub API Key: mock-github-key",
		"",
		"Model: gpt-4o",
		"",
	}, "\n")
	assert.Equal(t, expected, res, "Expected configuration to match")
}

func TestReal_PrintConfig_NoConfig(t *testing.T) {
	printer := output.NewMock()
	conf := config.NewMock()
	conf.Error = fmt.Errorf("no configuration found")
	raidy := &real{config: conf, printer: printer}

	err := raidy.PrintConfig()

	require.Error(t, err, "Expected no error when printing configuration with no config")
	res := printer.Captured()
	expected := strings.Join([]string{
		"Aidy Configuration:",
		"",
		"Error retrieving OpenAI API key: no configuration found",
		"",
		"Error retrieving Deepseek API key: no configuration found",
		"",
		"Error retrieving GitHub API key: no configuration found",
		"",
		"Error retrieving model: no configuration found",
		"",
	}, "\n")
	assert.Equal(t, expected, res, "Expected error message when no config is found")
}

func TestReal_Append(t *testing.T) {
	shell := executor.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell)}

	raidy.Append()

	expected := "git commit --amend --no-edit "
	require.Equal(t, 1, len(shell.Commands), "Expected number of commands to match")
	assert.Equal(t, expected, shell.Commands[0], "Expected command to match")
}

func TestReal_Heal(t *testing.T) {
	shell := executor.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell)}

	err := raidy.Heal()

	require.NoError(t, err, "Expected no error when healing")
	expected := []string{
		"git commit --amend -m feat(#41): current commit message",
	}
	for i, cmd := range expected {
		assert.Contains(t, shell.Commands[i], cmd, "Expected command '%s', got '%s'", cmd, shell.Commands[i])
	}
}

func TestReal_Heal_CantGetBranch(t *testing.T) {
	raidy := &real{git: git.NewMockWithError(fmt.Errorf("CurrentBranch method fails"))}

	err := raidy.Heal()

	require.Error(t, err, "Expected error when unable to get current branch")
	assert.Equal(t, "error getting branch name: CurrentBranch method fails", err.Error(), "Expected error message to match")
}

func TestReal_Heal_CantGetCommitMessage(t *testing.T) {
	raidy := &real{git: git.NewMockWithError(fmt.Errorf("CommitMessage method fails"))}

	err := raidy.Heal()

	require.Error(t, err, "Expected error when unable to get commit message")
	assert.Equal(t, "error getting current commit message: CommitMessage method fails", err.Error(), "Expected error message to match")
}

func TestReal_Heal_CantAmend(t *testing.T) {
	raidy := &real{git: git.NewMockWithError(fmt.Errorf("Amend method fails"))}

	err := raidy.Heal()

	require.Error(t, err, "Expected error when unable to amend commit")
	assert.Equal(t, "Amend method fails", err.Error(), "Expected error message to match")
}

func TestReal_CleanCache(t *testing.T) {
	tmp := t.TempDir()
	cache := filepath.Join(tmp, ".aidy")
	err := os.Mkdir(cache, 0755)
	require.NoError(t, err, "Failed to create .aidy directory")
	original, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	defer func() {
		_ = os.Chdir(original)
	}()
	err = os.Chdir(tmp)
	require.NoError(t, err, "Failed to change working directory")
	raidy := &real{}

	raidy.Clean()

	_, err = os.Stat(cache)
	assert.True(t, os.IsNotExist(err), ".aidy directory should be removed")
}

func TestReal_StartIssue(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh}

	err := raidy.StartIssue("42")

	assert.NoError(t, err, "Expected no error when starting issue")
	expected := []string{"git checkout -b 42-mock-branch-name"}
	assert.Equal(t, len(expected), len(shell.Commands), "number of commands should match")
	for i, expectedCommand := range expected {
		assert.Contains(t, shell.Commands[i], expectedCommand)
	}
}

func TestReal_StartIssueNoNumber(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh}

	err := raidy.StartIssue("")

	assert.Error(t, err, "expected error when no issue number is provided")
	assert.Contains(t, err.Error(), "error: no issue number provided")
}

func TestReal_StartIssueInvalidNumber(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh}

	err := raidy.StartIssue("invalid")

	assert.Error(t, err, "Expected error when an invalid issue number is provided")
	assert.Contains(t, err.Error(), "error: invalid issue number 'invalid'")
}

func TestReal_StartIssueBranchNameError(t *testing.T) {
	brain := ai.NewFailedMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh}

	err := raidy.StartIssue("42")

	assert.Error(t, err, "expected error when generating branch name fails")
	assert.Contains(t, err.Error(), "error generating branch name")
	assert.Contains(t, err.Error(), "failed to suggest branch")
}

func TestReal_StartIssueCheckoutError(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("error checking out branch")
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewMockAI(), github: github.NewMock()}

	err := raidy.StartIssue("42")

	assert.Error(t, err, "expected error when checking out branch fails")
	assert.Contains(t, err.Error(), "error checking out branch")
}

func TestReal_Squash(t *testing.T) {
	shell := executor.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewMockAI()}

	raidy.Squash()

	expected := []string{
		"git reset --soft refs/heads/main",
		"git add --all",
		"git commit -m feat(#41): no files changed",
		"git commit --amend -m feat(#41): current commit message",
	}
	for i, cmd := range expected {
		assert.Contains(t, shell.Commands[i], cmd, "Wrong command")
	}
}

func TestReal_PullRequest(t *testing.T) {
	mockGit := git.NewMock()
	mockAI := ai.NewMockAI()
	mockGithub := github.NewMock()
	out := output.NewMock()
	raidy := &real{git: mockGit, ai: mockAI, github: mockGithub, editor: out, cache: cache.NewMockAidyCache()}

	err := raidy.PullRequest()

	require.NoError(t, err, "Expected no error when creating pull request")
	output := out.Last()
	expected := "\ngh pr create --title \"Mock Title for 41\" --body \"Mock Body for 41\" --repo mock/remote"
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(output))
}

func TestReal_Commit(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	mgit := git.NewMockWithShell(shell)
	raidy := &real{git: mgit, ai: brain}

	err := raidy.Commit()

	require.NoError(t, err, "expected no error when committing changes")
	expected := []string{
		"git add --all",
		"git commit -m feat(#41): no files changed",
		"git commit --amend -m feat(#41): current commit message",
	}
	for i, cmd := range expected {
		assert.Contains(t, shell.Commands[i], cmd, "Expected command '%s', got '%s'", cmd, shell.Commands[i])
	}
}

func TestReal_Commit_CantGetCurrentBranch(t *testing.T) {
	mgit := git.NewMockWithError(fmt.Errorf("CurrentBranch method fails"))

	raidy := &real{git: mgit, ai: ai.NewMockAI()}

	err := raidy.Commit()
	require.Error(t, err, "expected error when unable to get current branch")
	assert.Equal(t, "error getting branch name: CurrentBranch method fails", err.Error(), "Expected error message to match")
}

func TestReal_Commit_CantGetCurrentDiff(t *testing.T) {
	mgit := git.NewMockWithError(fmt.Errorf("CurrentDiff method fails"))
	raidy := &real{git: mgit, ai: ai.NewMockAI()}

	err := raidy.Commit()

	require.Error(t, err, "expected error when unable to get current diff")
	assert.Equal(t, "error adding changes: CurrentDiff method fails", err.Error(), "Expected error message to match")
}

func TestReal_Commit_CantRunGit(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("git command failed")
	mgit := git.NewMockWithShell(shell)
	raidy := &real{git: mgit, ai: ai.NewMockAI()}

	err := raidy.Commit()

	require.Error(t, err, "expected error when git command fails")
	assert.Equal(t, "error adding changes: git command failed", err.Error(), "Expected error message to match")
}

func TestReal_Issue(t *testing.T) {
	userInput := "test input"
	out := output.NewMock()
	raidy := &real{ai: ai.NewMockAI(), github: github.NewMock(), editor: out, cache: cache.NewMockAidyCache()}

	err := raidy.Issue(userInput)

	require.NoError(t, err, "expected no error when creating issue")
	output := out.Last()
	expected := "\ngh issue create --title \"Mock Issue Title for test input\" --body \"Mock Issue Body for test input\" --label \"bug,documentation,question\" --repo mock/remote"
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(output))
}

func TestReal_Release_Success(t *testing.T) {
	mgit := git.NewMock()
	nobrain := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, editor: out}

	err := raidy.Release("minor", "origin")
	assert.NoError(t, err, "expected no error during release")
	expected := "git tag --cleanup=verbatim -a \"v2.1.0\" -m \""
	assert.Contains(t, out.Last(), expected, "expected release command to be generated")
}

func TestReal_Release_NoTags(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "absent"
	output := output.NewMock()
	mockGit := git.NewMockWithShell(shell)

	raidy := &real{git: mockGit, ai: ai.NewMockAI(), editor: output}

	err := raidy.Release("patch", "origin")

	require.NoError(t, err, "expected no error when releasing with no tags")
	expected := "git tag --cleanup=verbatim -a \"v0.0.1\" -m \""
	assert.Contains(t, output.Last(), expected, "expected release command to be generated with no tags")
}

func TestReal_ReleaseUnknownInterval(t *testing.T) {
	mockGit := git.NewMock()
	mockAI := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mockGit, ai: mockAI, editor: out}

	err := raidy.Release("", "origin")

	assert.EqualError(t, err, "failed to update version: 'unknown version step: '''", "expected error when no tags are present")
}

func TestReal_ReleaseTagFetchError(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("error fetching tags")
	mgit := git.NewMockWithShell(shell)
	nobrain := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, editor: out}

	err := raidy.Release("patch", "origin")

	assert.Error(t, err, "expected error when fetching tags fails")
	assert.Contains(t, err.Error(), "failed to get tags", "expected error message about fetching tags")
}

func TestReal_ReleaseNotesGenerationError(t *testing.T) {
	mgit := git.NewMock()
	nobrain := ai.NewFailedMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, editor: out}

	err := raidy.Release("major", "origin")

	assert.Error(t, err, "expected error when generating release notes fails")
	assert.Contains(t, err.Error(), "failed to generate release notes", "expected error message about release notes generation")
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
