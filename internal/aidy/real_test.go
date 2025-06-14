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
	"github.com/volodya-lombrozo/aidy/internal/log"
	"github.com/volodya-lombrozo/aidy/internal/output"
)

func TestReal_NewGitHub_RemovesSuccessfully(t *testing.T) {
	config := config.NewMock()
	config.Error = fmt.Errorf("mock error")
	git := git.NewMock()
	cache := cache.NewMockAidyCache()

	_, err := NewGitHub(git, config, cache)

	require.Error(t, err, "Expected no error when creating new GitHub instance")
	assert.Contains(t, err.Error(), "mock error", "Expected error message to match")
}

func TestReal_NewGitHub_CreatesSuccessfully(t *testing.T) {
	config := config.NewMock()
	git := git.NewMock()
	cache := cache.NewMockAidyCache()

	github, err := NewGitHub(git, config, cache)

	require.NoError(t, err, "Expected no error when creating new GitHub instance")
	assert.NotNil(t, github, "Expected GitHub instance to be initialized")
}

func TestReal_NewCache_NoFile(t *testing.T) {
	tmp := t.TempDir()
	filepath := filepath.Join(tmp, ".unexisting.json")
	git := git.NewMock()

	cache, err := NewCache(git, filepath)

	require.Error(t, err, "Expected error when cache file does not exist")
	assert.Contains(t, err.Error(), "can't open cache", "Expected error message to match")
	assert.Nil(t, cache, "Expected cache to be nil when file does not exist")
}

func TestReal_NewCache_CreatesSuccessfully(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".aidy")
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err, "Failed to create mock cache directory")
	path := filepath.Join(dir, "cache.json")
	err = os.WriteFile(path, []byte("{}"), 0666)
	require.NoError(t, err, "Failed to create mock cache file")
	git := git.NewMockWithDir(tmp)

	cache, err := NewCache(git, ".aidy/cache.json")

	require.NoError(t, err, "Expected no error when creating new cache")
	assert.NotNil(t, cache, "Expected cache to be initialized")
}

func TestReal_InitialisesAI_Mock(t *testing.T) {
	brain, err := Brain(true, false, config.NewMock())

	require.NoError(t, err, "Expected no error when initializing AI")
	assert.NotNil(t, brain, "Expected brain to be initialized")
}

func TestReal_InitialisesAI_OpenAI(t *testing.T) {
	conf := config.NewMock()
	conf.MockProvider = "openai"
	brain, err := Brain(false, false, conf)

	require.NoError(t, err, "Expected no error when initializing AI without cache")
	assert.NotNil(t, brain, "Expected brain to be initialized")
}

func TestReal_InitialisesAI_DeepSeek(t *testing.T) {
	conf := config.NewMock()
	conf.MockProvider = "deepseek"

	brain, err := Brain(false, false, conf)

	require.NoError(t, err, "Expected no error when initializing AI without cache")
	assert.NotNil(t, brain, "Expected brain to be initialized")
}

func TestReal_InitialisesAI_UnknownProvider(t *testing.T) {
	conf := config.NewMock()
	conf.MockProvider = "unknown"

	brain, err := Brain(false, false, conf)

	require.Error(t, err, "Expected error when initializing AI with unknown provider")
	assert.Nil(t, brain, "Expected brain to be nil when provider is unknown")
}

func TestReal_InitSummary_ErrorGettingProvider(t *testing.T) {
	conf := config.NewMock()
	conf.Error = fmt.Errorf("error getting provider")

	brain, err := Brain(false, false, conf)

	require.Error(t, err, "Expected error when getting provider fails")
	assert.Nil(t, brain, "Expected brain to be nil when getting provider fails")
	assert.Contains(t, err.Error(), "error getting provider", "Expected error message to contain 'error getting provider'")
}

func TestReal_InitSummary_CreateSummary(t *testing.T) {
	cache := cache.NewMockAidyCache()
	aidy := &real{ai: ai.NewMockAI(), git: git.NewMock(), config: config.NewMock(), cache: cache, printer: output.NewMock(), logger: log.NewMock()}
	tmp := t.TempDir()
	path := filepath.Join(tmp, "README.md")
	err := os.WriteFile(path, []byte("mock summary"), 0644)
	require.NoError(t, err, "Failed to create mock README.md file")

	err = aidy.InitSummary(true, path)

	require.NoError(t, err, "Expected no error when creating summary")
	summary, hash := cache.Summary()
	require.NoError(t, err, "Expected no error when retrieving summary from cache")
	assert.Equal(t, "summary: mock summary", summary, "Expected summary to be 'mock summary'")
	assert.Equal(t, "cad99a27bf4de48f", hash, "Expected hash to be 'mock-hash'")
}

func TestReal_InitSummary_AIError(t *testing.T) {
	cache := cache.NewMockAidyCache()
	brain := ai.NewFailedMockAI()
	aidy := &real{ai: brain, git: git.NewMock(), config: config.NewMock(), cache: cache, printer: output.NewMock(), logger: log.NewMock()}
	tmp := t.TempDir()
	path := filepath.Join(tmp, "README.md")
	err := os.WriteFile(path, []byte("mock summary"), 0644)
	require.NoError(t, err, "Failed to create mock README.md file")

	err = aidy.InitSummary(true, path)

	assert.Error(t, err, "expected error when AI fails to generate summary")
	assert.Equal(t, "error generating summary for README.md: failed to generate summary", err.Error(), "Expected error message to match")
}

func TestReal_InitSummary_CantFindReadme(t *testing.T) {
	aidy := &real{git: git.NewMock(), config: config.NewMock(), cache: cache.NewMockAidyCache(), printer: output.NewMock(), logger: log.NewMock()}

	err := aidy.InitSummary(true, "README.md")

	assert.Error(t, err, "expected error when readme.md is not found")
	assert.Contains(t, err.Error(), "can't read README.md file, because of", "Expected error message to match")
}

func TestReal_InitSummary_NotRequired(t *testing.T) {
	aidy := &real{git: git.NewMock(), config: config.NewMock(), cache: cache.NewMockAidyCache(), printer: output.NewMock()}

	err := aidy.InitSummary(false, "README.md")

	require.NoError(t, err, "Expected no error when summary is not required")
}

func TestReal_CheckGitInstalled_Successfully(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "git version 2.34.1"
	aidy := &real{git: git.NewMockWithShell(shell), config: config.NewMock(), cache: cache.NewMockAidyCache(), printer: output.NewMock()}

	err := aidy.CheckGitInstalled()

	require.NoError(t, err, "Expected no error when git is installed")
}

func TestReal_CheckGitInstalled_NotInstalled(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "git is not installed on the system"
	aidy := &real{git: git.NewMockWithShell(shell), config: config.NewMock(), cache: cache.NewMockAidyCache(), printer: output.NewMock()}

	err := aidy.CheckGitInstalled()

	assert.Error(t, err)
	assert.Equal(t, "git is not installed on the system", err.Error(), "Expected error message to match")
}

func TestReal_CheckGitInstalled_Error(t *testing.T) {
	aidy := &real{git: git.NewMockWithError(fmt.Errorf("git not found")), config: config.NewMock(), cache: cache.NewMockAidyCache(), printer: output.NewMock()}

	err := aidy.CheckGitInstalled()

	require.Error(t, err, "Expected error when git is not installed")
	assert.Equal(t, "can't determine whether git is installed or not, because of 'git not found'", err.Error(), "Expected error message to match")
}

func TestReal_SetTarget_IfOnlyOne(t *testing.T) {
	cache := cache.NewMockAidyCache()
	aidy := &real{git: git.NewMock(), config: config.NewMock(), cache: cache, printer: output.NewMock(), logger: log.NewMock()}

	err := aidy.SetTarget()

	require.NoError(t, err, "Expected no error when setting target with only one remote")
	actual := cache.Remote()
	assert.Equal(t, "mock/remote", actual, "Expected remote to be set to 'mock/remote'")
}

func TestReal_SetTarget_NoRepos(t *testing.T) {
	cache := cache.NewMockAidyCache()
	cache.WithRemote("")
	aidy := &real{git: git.NewMock(), config: config.NewMock(), cache: cache, printer: output.NewMock(), logger: log.NewMock()}

	err := aidy.SetTarget()

	require.Error(t, err, "Expected error when no remotes are available")
	assert.Equal(t, "no remote repositories found, please set one", err.Error(), "Expected error message to match")
}

func TestReal_SetTarget_RealGitOutput(t *testing.T) {
	cache := cache.NewMockAidyCache()
	cache.WithRemote("")
	shell := executor.NewMock()
	shell.Output = "upstream        git@github.com:yegor256/jaxec.git (fetch)"
	git := git.NewMockWithShell(shell)
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe for input")
	_, err = w.WriteString("1")
	require.NoError(t, err, "Failed to write to pipe")
	aidy := &real{in: r, git: git, config: config.NewMock(), cache: cache, printer: output.NewMock(), logger: log.NewMock()}
	err = w.Close()
	require.NoError(t, err, "Failed to close pipe for input")

	err = aidy.SetTarget()

	require.NoError(t, err, "Expected no error when setting target with multiple remotes")
	assert.Equal(t, "yegor256/jaxec", cache.Remote(), "Expected remote to be set to 'cqfn/kaicode.github.io'")
}

func TestReal_SetTarget_MultipleRemotes(t *testing.T) {
	cache := cache.NewMockAidyCache()
	cache.WithRemote("")
	shell := executor.NewMock()
	shell.Output = "https://github.com/cqfn/kaicode.github.io.git\nhttps://github.com/volodya-lombrozo/aidy.git\n"
	git := git.NewMockWithShell(shell)
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe for input")
	_, err = w.WriteString("1")
	require.NoError(t, err, "Failed to write to pipe")
	aidy := &real{in: r, git: git, config: config.NewMock(), cache: cache, printer: output.NewMock(), logger: log.NewMock()}
	err = w.Close()
	require.NoError(t, err, "Failed to close pipe for input")

	err = aidy.SetTarget()

	require.NoError(t, err, "Expected no error when setting target with multiple remotes")
	assert.Equal(t, "cqfn/kaicode.github.io", cache.Remote(), "Expected remote to be set to 'cqfn/kaicode.github.io'")
}

func TestReal_SetTarget_MultipleRemotes_WrongChoice(t *testing.T) {
	cache := cache.NewMockAidyCache()
	cache.WithRemote("")
	shell := executor.NewMock()
	shell.Output = "https://github.com/cqfn/kaicode.github.io.git\nhttps://github.com/volodya-lombrozo/aidy.git\n"
	git := git.NewMockWithShell(shell)
	r, w, err := os.Pipe()
	require.NoError(t, err, "Failed to create pipe for input")
	_, err = w.WriteString("3")
	require.NoError(t, err, "Failed to write to pipe")
	aidy := &real{in: r, git: git, config: config.NewMock(), cache: cache, printer: output.NewMock(), logger: log.NewMock()}
	err = w.Close()
	require.NoError(t, err, "Failed to close pipe for input")

	err = aidy.SetTarget()

	assert.Error(t, err, "Expected error when choosing a remote that does not exist")
	assert.Equal(t, "invalid choice: <nil>", err.Error(), "Expected error message to match")
}

func TestReal_SetTarget_GitError(t *testing.T) {
	cache := cache.NewMockAidyCache()
	cache.WithRemote("")
	aidy := &real{
		git:     git.NewMockWithError(fmt.Errorf("mock error")),
		config:  config.NewMock(),
		cache:   cache,
		printer: output.NewMock(),
		logger:  log.NewMock(),
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
		logger:  log.NewMock(),
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
	expected := "diff with the base branch:\nmock-diff\n"
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
		"aidy configuration:",
		"AI provider: openai",
		"model: gpt-4o",
		"AI API token: ******oken",
		"GitHub API key: ***********-key",
	}, "\n")
	assert.Equal(t, expected, res, "Expected configuration to match")
}

func TestReal_PrintConfig_ShortKeys(t *testing.T) {
	printer := output.NewMock()
	conf := config.NewMock()
	conf.MockToken = "short"
	raidy := &real{config: conf, printer: printer}

	err := raidy.PrintConfig()

	require.NoError(t, err, "Expected no error when printing configuration")
	res := printer.Captured()
	assert.Contains(t, res, "aidy configuration:")
	assert.Contains(t, res, "AI API token: *hort")
}

func TestReal_PrintConfig_TooShort(t *testing.T) {
	printer := output.NewMock()
	conf := config.NewMock()
	conf.MockToken = "123"
	raidy := &real{config: conf, printer: printer}

	err := raidy.PrintConfig()

	require.NoError(t, err, "Expected no error when printing configuration with short key")
	res := printer.Captured()
	assert.Contains(t, res, "aidy configuration:")
	assert.Contains(t, res, "API token: ***")
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
		"aidy configuration:",
		"error retrieving AI provider: no configuration found",
		"error retrieving model: no configuration found",
		"error retrieving AI token: no configuration found",
		"error retrieving GitHub API key: no configuration found",
	}, "\n")
	assert.Equal(t, expected, res, "Expected error message when no config is found")
}

func TestReal_Append(t *testing.T) {
	shell := executor.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), logger: log.Get()}

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
	raidy := &real{logger: log.Get()}

	raidy.Clean()

	_, err = os.Stat(cache)
	assert.True(t, os.IsNotExist(err), ".aidy directory should be removed")
}

func TestReal_StartIssue(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh, logger: log.Get()}

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
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh, logger: log.Get()}

	err := raidy.StartIssue("")

	assert.Error(t, err, "expected error when no issue number is provided")
	assert.Contains(t, err.Error(), "error: no issue number provided")
}

func TestReal_StartIssueInvalidNumber(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh, logger: log.Get()}

	err := raidy.StartIssue("invalid")

	assert.Error(t, err, "Expected error when an invalid issue number is provided")
	assert.Contains(t, err.Error(), "error: invalid issue number 'invalid'")
}

func TestReal_StartIssueBranchNameError(t *testing.T) {
	brain := ai.NewFailedMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh, logger: log.Get()}

	err := raidy.StartIssue("42")

	assert.Error(t, err, "expected error when generating branch name fails")
	assert.Contains(t, err.Error(), "error generating branch name")
	assert.Contains(t, err.Error(), "failed to suggest branch")
}

func TestReal_StartIssueCheckoutError(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("error checking out branch")
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewMockAI(), github: github.NewMock(), logger: log.Get()}

	err := raidy.StartIssue("42")

	assert.Error(t, err, "expected error when checking out branch fails")
	assert.Contains(t, err.Error(), "error checking out branch")
}

func TestReal_Squash(t *testing.T) {
	shell := executor.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewMockAI(), logger: log.Get()}

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
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), editor: out, cache: cache.NewMockAidyCache(), logger: log.Get()}

	err := raidy.PullRequest()

	require.NoError(t, err, "expected no error when creating pull request")
	output := out.Last()
	assert.Contains(t, output, "gh pr create", "Expected output to contain 'gh pr create'")
	assert.Contains(t, output, "--title \"mock title for '41' with issue #mock description for issue '#41' and summary: mock summary\"", "Expected output to contain title")
}

func TestReal_PullRequest_IssueNotFound(t *testing.T) {
	github := github.NewMock()
	github.Error = fmt.Errorf("issue not found")
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github, editor: out, cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.PullRequest()

	require.NoError(t, err, "Expected no error when creating pull request with issue not found")
	output := out.Last()
	assert.Contains(t, output, "gh pr create")
	assert.Contains(t, output, "with issue #not-found")
}

func TestReal_Commit(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	mgit := git.NewMockWithShell(shell)
	raidy := &real{git: mgit, ai: brain, logger: log.Get()}

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

	raidy := &real{git: mgit, ai: ai.NewMockAI(), logger: log.Get()}

	err := raidy.Commit()
	require.Error(t, err, "expected error when unable to get current branch")
	assert.Equal(t, "error getting branch name: CurrentBranch method fails", err.Error(), "Expected error message to match")
}

func TestReal_Commit_CantGetCurrentDiff(t *testing.T) {
	mgit := git.NewMockWithError(fmt.Errorf("CurrentDiff method fails"))
	raidy := &real{git: mgit, ai: ai.NewMockAI(), logger: log.Get()}

	err := raidy.Commit()

	require.Error(t, err, "expected error when unable to get current diff")
	assert.Equal(t, "error adding changes: CurrentDiff method fails", err.Error(), "Expected error message to match")
}

func TestReal_Commit_CantRunGit(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("git command failed")
	mgit := git.NewMockWithShell(shell)
	raidy := &real{git: mgit, ai: ai.NewMockAI(), logger: log.Get()}

	err := raidy.Commit()

	require.Error(t, err, "expected error when git command fails")
	assert.Equal(t, "error adding changes: git command failed", err.Error(), "Expected error message to match")
}

func TestReal_Issue(t *testing.T) {
	userInput := "test input"
	out := output.NewMock()
	raidy := &real{ai: ai.NewMockAI(), github: github.NewMock(), editor: out, cache: cache.NewMockAidyCache(), logger: log.Get()}

	err := raidy.Issue(userInput)

	require.NoError(t, err, "expected no error when creating issue")
	output := out.Last()
	assert.Contains(t, output, "gh issue create")
	assert.Contains(t, output, "--title \"mock issue title for 'test input' with summary: mock summary\"")
	assert.Contains(t, output, "--body \"mock issue body for 'test input' with summary: mock summary\"")
	assert.Contains(t, output, "--label \"bug,documentation,question\"")
}

func TestReal_Release_Success(t *testing.T) {
	mgit := git.NewMock()
	nobrain := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, editor: out, logger: log.NewMock()}

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

	raidy := &real{git: mockGit, ai: ai.NewMockAI(), editor: output, logger: log.NewMock()}

	err := raidy.Release("patch", "origin")

	require.NoError(t, err, "expected no error when releasing with no tags")
	expected := "git tag --cleanup=verbatim -a \"v0.0.1\" -m \""
	assert.Contains(t, output.Last(), expected, "expected release command to be generated with no tags")
}

func TestReal_ReleaseUnknownInterval(t *testing.T) {
	mockGit := git.NewMock()
	mockAI := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mockGit, ai: mockAI, editor: out, logger: log.NewMock()}

	err := raidy.Release("", "origin")

	assert.EqualError(t, err, "failed to update version: 'unknown version step: '''", "expected error when no tags are present")
}

func TestReal_Release_TagFetchError(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("error fetching tags")
	mgit := git.NewMockWithShell(shell)
	nobrain := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, editor: out, logger: log.NewMock()}

	err := raidy.Release("patch", "origin")

	assert.Error(t, err, "expected error when fetching tags fails")
	assert.Contains(t, err.Error(), "failed to get tags", "expected error message about fetching tags")
}

func TestReal_Release_NotesGenerationError(t *testing.T) {
	mgit := git.NewMock()
	nobrain := ai.NewFailedMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, editor: out, logger: log.NewMock()}

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

func TestInitLogger_DebugMode(t *testing.T) {
	InitLogger(false, true)

	logger := log.Get()

	assert.NotNil(t, logger, "Expected logger to be initialized")
	assert.IsType(t, &log.Short{}, logger, "Expected logger to be of type Short")
}

func TestInitLogger_SilentMode(t *testing.T) {
	InitLogger(true, false)

	logger := log.Get()

	assert.NotNil(t, logger, "Expected logger to be initialized")
	assert.IsType(t, &log.Silent{}, logger, "Expected logger to be of type Silent")
}

func TestInitLogger_DefaultMode(t *testing.T) {
	InitLogger(false, false)

	logger := log.Get()

	assert.NotNil(t, logger, "Expected logger to be initialized")
	assert.IsType(t, &log.Short{}, logger, "Expected logger to be of type Short")
}
