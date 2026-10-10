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
	"github.com/volodya-lombrozo/aidy/internal/gitlab"
	"github.com/volodya-lombrozo/aidy/internal/jira"
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
	brain, err := Brain(true, false, config.NewMock(), "en")

	require.NoError(t, err, "Expected no error when initializing AI")
	assert.NotNil(t, brain, "Expected brain to be initialized")
}

func TestReal_InitialisesAI_OpenAI(t *testing.T) {
	conf := config.NewMock()
	conf.MockProvider = "openai"
	brain, err := Brain(false, false, conf, "en")

	require.NoError(t, err, "Expected no error when initializing AI without cache")
	assert.NotNil(t, brain, "Expected brain to be initialized")
}

func TestReal_InitialisesAI_DeepSeek(t *testing.T) {
	conf := config.NewMock()
	conf.MockProvider = "deepseek"

	brain, err := Brain(false, false, conf, "en")

	require.NoError(t, err, "Expected no error when initializing AI without cache")
	assert.NotNil(t, brain, "Expected brain to be initialized")
}

func TestReal_InitialisesAI_UnknownProvider(t *testing.T) {
	conf := config.NewMock()
	conf.MockProvider = "unknown"

	brain, err := Brain(false, false, conf, "en")

	require.Error(t, err, "Expected error when initializing AI with unknown provider")
	assert.Nil(t, brain, "Expected brain to be nil when provider is unknown")
}

func TestReal_InitSummary_ErrorGettingProvider(t *testing.T) {
	conf := config.NewMock()
	conf.Error = fmt.Errorf("error getting provider")

	brain, err := Brain(false, false, conf, "en")

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
	raidy := &real{git: git.NewMockWithShell(shell), logger: log.Default()}

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
	raidy := &real{logger: log.Default()}

	raidy.Clean()

	_, err = os.Stat(cache)
	assert.True(t, os.IsNotExist(err), ".aidy directory should be removed")
}

func TestReal_StartIssue(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh, cache: cache.NewMockAidyCache(), logger: log.Default()}

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
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.StartIssue("")

	assert.Error(t, err, "expected error when no issue number is provided")
	assert.Contains(t, err.Error(), "error: no issue number provided")
}

func TestReal_StartIssueInvalidNumber(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.StartIssue("invalid")

	assert.Error(t, err, "Expected error when an invalid issue number is provided")
	assert.Contains(t, err.Error(), "error: invalid issue number 'invalid'")
}

func TestReal_StartIssueBranchNameError(t *testing.T) {
	brain := ai.NewFailedMockAI()
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: brain, github: gh, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.StartIssue("42")

	assert.Error(t, err, "expected error when generating branch name fails")
	assert.Contains(t, err.Error(), "error generating branch name")
	assert.Contains(t, err.Error(), "failed to suggest branch")
}

func TestReal_StartIssueCheckoutError(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("error checking out branch")
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewMockAI(), github: github.NewMock(), cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.StartIssue("42")

	assert.Error(t, err, "expected error when checking out branch fails")
	assert.Contains(t, err.Error(), "error checking out branch")
}

func TestReal_Squash(t *testing.T) {
	shell := executor.NewMock()
	raidy := &real{github: github.NewMock(), git: git.NewMockWithShell(shell), ai: ai.NewMockAI(), cache: cache.NewMockAidyCache(), logger: log.Default()}

	raidy.Squash(true)

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
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.PullRequest(false, "", false, "")

	require.NoError(t, err, "expected no error when creating pull request")
	output := out.Last()
	assert.Contains(t, output, "gh pr create", "Expected output to contain 'gh pr create'")
	assert.Contains(t, output, "--title \"mock title for '#41' with issue #mock description for issue '#41' and summary: mock summary\"", "Expected output to contain title")
}

func TestReal_PullRequest_Target(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.PullRequest(false, "develop", false, "")

	require.NoError(t, err, "expected no error when creating pull request with target branch")
	output := out.Last()
	assert.Contains(t, output, "--base develop", "Expected output to contain target branch")
	assert.Contains(t, output, "mock-diff for develop", "Expected the diff to be taken against the target branch")
}

func TestReal_PullRequest_TargetAndSource(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.PullRequest(false, "develop", false, "feature-x")

	require.NoError(t, err, "expected no error when creating pull request with target and source branches")
	output := out.Last()
	assert.Contains(t, output, "--base develop", "Expected output to contain target branch")
	assert.Contains(t, output, "mock-diff for develop feature-x", "Expected the diff to be taken between the target and the source branches")
}

func TestReal_PullRequest_Source(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.PullRequest(false, "", false, "feature-x")

	require.NoError(t, err, "expected no error when creating pull request with source branch")
	assert.Contains(t, out.Last(), "mock-diff for feature-x", "Expected the diff to be taken from the source branch")
}

func TestReal_PullRequest_IssueNotFound(t *testing.T) {
	github := github.NewMock()
	github.Error = fmt.Errorf("issue not found")
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github, reviewer: out, cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.PullRequest(false, "", false, "")

	require.NoError(t, err, "Expected no error when creating pull request with issue not found")
	output := out.Last()
	assert.Contains(t, output, "gh pr create")
	assert.Contains(t, output, "with issue #not-found")
	assert.Contains(t, output, "Related to #")
}

func TestReal_PullRequest_Fixes(t *testing.T) {
	github := github.NewMock()
	github.Error = fmt.Errorf("issue not found")
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github, reviewer: out, cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.PullRequest(true, "", false, "")

	require.NoError(t, err, "Expected no error when creating pull request with issue not found")
	output := out.Last()
	assert.Contains(t, output, "Fixes #")
}

func TestReal_PullRequest_Regenerates(t *testing.T) {
	out := output.NewMock()
	out.Generations = 3
	hub := github.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: hub, reviewer: out, cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.PullRequest(false, "", false, "")

	require.NoError(t, err, "expected no error when regenerating a pull request")
	assert.Equal(t, 1, hub.Descriptions, "expected the issue to be fetched once, no matter how many commands are generated")
	assert.Contains(t, out.Last(), "gh pr create", "expected the last generated command to be the one shown")
}

// headersAI generates a body with markdown section headers, the way a model
// sometimes does despite being told not to.
type headersAI struct {
	ai.AI
}

func (h *headersAI) PrBody(diff string, issue string, summary string) (string, error) {
	return "## Description\n\nEscapes double quotes.\n\n## Testing\n\nNew tests added.", nil
}

func TestReal_PullRequest_StripsHeaders(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: &headersAI{ai.NewMockAI()}, github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.PullRequest(false, "", false, "")

	require.NoError(t, err, "expected no error when creating pull request")
	result := out.Last()
	assert.NotContains(t, result, "##", "Expected markdown headers to be stripped from the body")
	assert.Contains(t, result, "Escapes double quotes.", "Expected the prose to survive header stripping")
	assert.Contains(t, result, "New tests added.", "Expected the prose to survive header stripping")
}

func TestReal_PullRequest_Duplicate(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.PullRequest(false, "develop", true, "")

	require.NoError(t, err, "expected no error when duplicating a pull request")
	output := out.Last()
	assert.Contains(t, output, "gh pr create", "Expected output to contain 'gh pr create'")
	assert.Contains(t, output, "--base develop", "Expected output to contain target branch")
	assert.Contains(t, output, "mock title for branch '41_working_branch'", "Expected output to reuse the current branch's pull request title")
	assert.Contains(t, output, "mock body for branch '41_working_branch'", "Expected output to reuse the current branch's pull request body")
}

func TestReal_PullRequest_Duplicate_Source(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.PullRequest(false, "develop", true, "feature-x")

	require.NoError(t, err, "expected no error when duplicating a pull request from a given source branch")
	output := out.Last()
	assert.Contains(t, output, "--base develop", "Expected output to contain target branch")
	assert.Contains(t, output, "mock title for branch 'feature-x'", "Expected output to reuse the source branch's pull request title")
	assert.Contains(t, output, "mock body for branch 'feature-x'", "Expected output to reuse the source branch's pull request body")
}

func TestReal_PullRequest_Duplicate_NoTarget(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.PullRequest(false, "", true, "")

	require.Error(t, err, "expected an error when duplicating a pull request without a target branch")
	assert.Contains(t, err.Error(), "--duplicate requires --target")
}

func TestReal_PullRequest_Duplicate_NotFound(t *testing.T) {
	github := github.NewMock()
	github.Error = fmt.Errorf("no pull request found for branch 'feature'")
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github, reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.PullRequest(false, "develop", true, "")

	require.Error(t, err, "expected an error when there is no pull request to duplicate")
	assert.Contains(t, err.Error(), "error finding an existing pull request to duplicate")
}

func TestReal_MergeRequest(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.MergeRequest(false, "", false, "")

	require.NoError(t, err, "expected no error when creating merge request")
	result := out.Last()
	assert.Contains(t, result, "glab mr create", "Expected output to contain 'glab mr create'")
	assert.Contains(t, result, "Related to #")
}

func TestReal_MergeRequest_Target(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.MergeRequest(false, "develop", false, "")

	require.NoError(t, err, "expected no error when creating merge request with target branch")
	result := out.Last()
	assert.Contains(t, result, "--target-branch develop", "Expected output to contain target branch")
	assert.Contains(t, result, "mock-diff for develop", "Expected the diff to be taken against the target branch")
}

func TestReal_MergeRequest_TargetAndSource(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.MergeRequest(false, "develop", false, "feature-x")

	require.NoError(t, err, "expected no error when creating merge request with target and source branches")
	result := out.Last()
	assert.Contains(t, result, "--target-branch develop", "Expected output to contain target branch")
	assert.Contains(t, result, "mock-diff for develop feature-x", "Expected the diff to be taken between the target and the source branches")
}

func TestReal_MergeRequest_Source(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.MergeRequest(false, "", false, "feature-x")

	require.NoError(t, err, "expected no error when creating merge request with source branch")
	assert.Contains(t, out.Last(), "mock-diff for feature-x", "Expected the diff to be taken from the source branch")
}

func TestReal_MergeRequest_Fixes(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.MergeRequest(true, "", false, "")

	require.NoError(t, err, "expected no error when creating merge request with fixes")
	result := out.Last()
	assert.Contains(t, result, "glab mr create")
	assert.Contains(t, result, "Closes #")
}

func TestReal_MergeRequest_StripsHeaders(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: &headersAI{ai.NewMockAI()}, github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.MergeRequest(false, "", false, "")

	require.NoError(t, err, "expected no error when creating merge request")
	result := out.Last()
	assert.NotContains(t, result, "##", "Expected markdown headers to be stripped from the description")
	assert.Contains(t, result, "Escapes double quotes.", "Expected the prose to survive header stripping")
}

func TestReal_MergeRequest_Duplicate(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), gitlab: gitlab.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.MergeRequest(false, "develop", true, "")

	require.NoError(t, err, "expected no error when duplicating a merge request")
	result := out.Last()
	assert.Contains(t, result, "glab mr create", "Expected output to contain 'glab mr create'")
	assert.Contains(t, result, "--target-branch develop", "Expected output to contain target branch")
	assert.Contains(t, result, "mock title for branch '41_working_branch'", "Expected output to reuse the current branch's merge request title")
	assert.Contains(t, result, "mock body for branch '41_working_branch'", "Expected output to reuse the current branch's merge request body")
}

func TestReal_MergeRequest_Duplicate_Source(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), gitlab: gitlab.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.MergeRequest(false, "develop", true, "feature-x")

	require.NoError(t, err, "expected no error when duplicating a merge request from a given source branch")
	result := out.Last()
	assert.Contains(t, result, "--target-branch develop", "Expected output to contain target branch")
	assert.Contains(t, result, "mock title for branch 'feature-x'", "Expected output to reuse the source branch's merge request title")
	assert.Contains(t, result, "mock body for branch 'feature-x'", "Expected output to reuse the source branch's merge request body")
}

func TestReal_MergeRequest_Duplicate_NoTarget(t *testing.T) {
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), gitlab: gitlab.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.MergeRequest(false, "", true, "")

	require.Error(t, err, "expected an error when duplicating a merge request without a target branch")
	assert.Contains(t, err.Error(), "--duplicate requires --target")
}

func TestReal_MergeRequest_Duplicate_NotFound(t *testing.T) {
	gl := gitlab.NewMock()
	gl.Error = fmt.Errorf("no merge request found for branch 'feature'")
	out := output.NewMock()
	raidy := &real{git: git.NewMock(), ai: ai.NewMockAI(), github: github.NewMock(), gitlab: gl, reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.MergeRequest(false, "develop", true, "")

	require.Error(t, err, "expected an error when there is no merge request to duplicate")
	assert.Contains(t, err.Error(), "error finding an existing merge request to duplicate")
}

func TestReal_Commit(t *testing.T) {
	brain := ai.NewMockAI()
	shell := executor.NewMock()
	mgit := git.NewMockWithShell(shell)
	raidy := &real{git: mgit, ai: brain, logger: log.Default(), github: github.NewMock(), cache: cache.NewMockAidyCache()}

	err := raidy.Commit(true)

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

	raidy := &real{git: mgit, ai: ai.NewMockAI(), logger: log.Default(), cache: cache.NewMockAidyCache()}

	err := raidy.Commit(true)
	require.Error(t, err, "expected error when unable to get current branch")
	assert.Equal(t, "error getting branch name: CurrentBranch method fails", err.Error(), "Expected error message to match")
}

func TestReal_Commit_CantGetCurrentDiff(t *testing.T) {
	mgit := git.NewMockWithError(fmt.Errorf("CurrentDiff method fails"))
	raidy := &real{git: mgit, ai: ai.NewMockAI(), logger: log.Default(), cache: cache.NewMockAidyCache(), github: github.NewMock()}

	err := raidy.Commit(true)

	require.Error(t, err, "expected error when unable to get current diff")
	assert.Equal(t, "error adding changes: CurrentDiff method fails", err.Error(), "Expected error message to match")
}

func TestReal_Commit_CantRunGit(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("git command failed")
	mgit := git.NewMockWithShell(shell)
	raidy := &real{git: mgit, ai: ai.NewMockAI(), logger: log.Default(), cache: cache.NewMockAidyCache(), github: github.NewMock()}

	err := raidy.Commit(true)

	require.Error(t, err, "expected error when git command fails")
	assert.Equal(t, "error adding changes: git command failed", err.Error(), "Expected error message to match")
}

func TestReal_Issue(t *testing.T) {
	userInput := "test input"
	out := output.NewMock()
	raidy := &real{ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.Issue(userInput)

	require.NoError(t, err, "expected no error when creating issue")
	output := out.Last()
	assert.Contains(t, output, "gh issue create")
	assert.Contains(t, output, "--title \"mock issue title for 'test input' with summary: mock summary\"")
	assert.Contains(t, output, "--body \"mock issue body for 'test input' with summary: mock summary\"")
	assert.Contains(t, output, "--label \"bug,documentation,question\"")
}

func TestReal_Issue_EscapesDoubleQuotes(t *testing.T) {
	userInput := `task with "quoted" input`
	out := output.NewMock()
	raidy := &real{ai: ai.NewMockAI(), github: github.NewMock(), reviewer: out, cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.Issue(userInput)

	require.NoError(t, err, "expected no error when creating issue")
	output := out.Last()
	assert.Contains(t, output, `\"quoted\"`, "expected embedded double quotes to be escaped so the generated command stays valid")
	assert.NotContains(t, output, `"quoted"`, "unescaped double quotes would break the generated shell command")
}

func TestReal_Release_Success(t *testing.T) {
	mgit := git.NewMock()
	nobrain := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, reviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", false)
	assert.NoError(t, err, "expected no error during release")
	expected := "git tag --cleanup=verbatim -a \"v2.1.0\" -m \""
	assert.Contains(t, out.Last(), expected, "expected release command to be generated")
}

func TestReal_Release_Regenerates(t *testing.T) {
	shell := executor.NewMock()
	out := output.NewMock()
	out.Generations = 3
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewMockAI(), reviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", false)

	require.NoError(t, err, "expected no error when regenerating release notes")
	logs := 0
	for _, command := range shell.Commands {
		if strings.HasPrefix(command, "git log") {
			logs++
		}
	}
	assert.Equal(t, 1, logs, "expected the git log to be read once, no matter how many notes are generated")
	assert.Contains(t, out.Last(), "git tag --cleanup=verbatim -a \"v2.1.0\"", "expected the last generated command to be the one shown")
}

func TestReal_Release_SaveNotes_Regenerates(t *testing.T) {
	tmp := t.TempDir()
	shell := executor.NewMock()
	shell.Output = "https://github.com/volodya-lombrozo/aidy.git"
	out := output.NewMock()
	out.Generations = 2
	raidy := &real{git: git.NewMockWithDirAndShell(tmp, shell), ai: ai.NewMockAI(), reviewer: out, textreviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", true)

	require.NoError(t, err, "expected no error when regenerating release notes")
	notes, rerr := os.ReadFile(filepath.Join(tmp, ".github", "release-notes", "v2.1.0.md"))
	require.NoError(t, rerr, "expected release notes file to be written")
	assert.Contains(t, out.Last(), "-m \""+string(notes)+"\"", "expected the saved notes and the tag message to be the same version")
}

func TestReal_Release_NoTags_Patch(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "absent"
	output := output.NewMock()
	mockGit := git.NewMockWithShell(shell)

	raidy := &real{git: mockGit, ai: ai.NewMockAI(), reviewer: output, logger: log.NewMock()}

	err := raidy.Release("patch", "origin", false)

	require.NoError(t, err, "expected no error when releasing with no tags")
	expected := "git tag --cleanup=verbatim -a \"v0.0.1\" -m \""
	assert.Contains(t, output.Last(), expected, "expected release command to be generated with no tags")
}

func TestReal_Release_NoTags_Minor(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "absent"
	output := output.NewMock()
	mockGit := git.NewMockWithShell(shell)

	raidy := &real{git: mockGit, ai: ai.NewMockAI(), reviewer: output, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", false)

	require.NoError(t, err, "expected no error when releasing with no tags")
	expected := "git tag --cleanup=verbatim -a \"v0.1.0\" -m \""
	assert.Contains(t, output.Last(), expected, "expected release command to be generated with no tags")
}

func TestReal_Release_NoTags_Major(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "absent"
	output := output.NewMock()
	mockGit := git.NewMockWithShell(shell)

	raidy := &real{git: mockGit, ai: ai.NewMockAI(), reviewer: output, logger: log.NewMock()}

	err := raidy.Release("major", "origin", false)

	require.NoError(t, err, "expected no error when releasing with no tags")
	expected := "git tag --cleanup=verbatim -a \"v1.0.0\" -m \""
	assert.Contains(t, output.Last(), expected, "expected release command to be generated with no tags")
}

func TestReal_ReleaseUnknownInterval(t *testing.T) {
	mockGit := git.NewMock()
	mockAI := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mockGit, ai: mockAI, reviewer: out, logger: log.NewMock()}

	err := raidy.Release("", "origin", false)

	assert.EqualError(t, err, "failed to update version: 'unknown version step: '''", "expected error when no tags are present")
}

func TestReal_Release_TagFetchError(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("error fetching tags")
	mgit := git.NewMockWithShell(shell)
	nobrain := ai.NewMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, reviewer: out, logger: log.NewMock()}

	err := raidy.Release("patch", "origin", false)

	assert.Error(t, err, "expected error when fetching tags fails")
	assert.Contains(t, err.Error(), "failed to get tags", "expected error message about fetching tags")
}

func TestReal_Release_NotesGenerationError(t *testing.T) {
	mgit := git.NewMock()
	nobrain := ai.NewFailedMockAI()
	out := output.NewMock()
	raidy := &real{git: mgit, ai: nobrain, reviewer: out, logger: log.NewMock()}

	err := raidy.Release("major", "origin", false)

	assert.Error(t, err, "expected error when generating release notes fails")
	assert.Contains(t, err.Error(), "failed to generate release notes", "expected error message about release notes generation")
}

func TestReal_Release_SaveNotes_GitHub(t *testing.T) {
	tmp := t.TempDir()
	shell := executor.NewMock()
	shell.Output = "https://github.com/volodya-lombrozo/aidy.git"
	mgit := git.NewMockWithDirAndShell(tmp, shell)
	out := output.NewMock()
	raidy := &real{git: mgit, ai: ai.NewMockAI(), reviewer: out, textreviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", true)

	require.NoError(t, err, "expected no error during release")
	path := filepath.Join(tmp, ".github", "release-notes", "v2.1.0.md")
	notes, rerr := os.ReadFile(path)
	require.NoError(t, rerr, "expected release notes file to be written under .github")
	assert.Contains(t, string(notes), "Mock Release Notes", "expected release notes file to contain generated notes")
	commands := strings.Join(shell.Commands, "\n")
	assert.Contains(t, commands, "git add "+path, "expected release notes file to be staged")
	assert.Contains(t, commands, "git commit -m chore: add release notes for v2.1.0", "expected release notes to be committed")
}

func TestReal_Release_SaveNotes_UsesReviewedText(t *testing.T) {
	tmp := t.TempDir()
	shell := executor.NewMock()
	shell.Output = "https://github.com/volodya-lombrozo/aidy.git"
	mgit := git.NewMockWithDirAndShell(tmp, shell)
	out := output.NewMock()
	out.ReviewText = "edited release notes"
	raidy := &real{git: mgit, ai: ai.NewMockAI(), reviewer: out, textreviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", true)

	require.NoError(t, err, "expected no error during release")
	path := filepath.Join(tmp, ".github", "release-notes", "v2.1.0.md")
	notes, rerr := os.ReadFile(path)
	require.NoError(t, rerr, "expected release notes file to be written")
	assert.Equal(t, "edited release notes", string(notes), "expected the reviewed/edited text to be saved, not the original AI output")
	assert.Contains(t, out.Last(), "-m \"edited release notes\"", "expected the tag message to use the reviewed text")
}

func TestReal_Release_SaveNotes_Canceled(t *testing.T) {
	tmp := t.TempDir()
	shell := executor.NewMock()
	shell.Output = "https://github.com/volodya-lombrozo/aidy.git"
	mgit := git.NewMockWithDirAndShell(tmp, shell)
	out := output.NewMock()
	out.ReviewErr = output.ErrCanceled
	raidy := &real{git: mgit, ai: ai.NewMockAI(), reviewer: out, textreviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", true)

	require.NoError(t, err, "expected no error when the notes review is canceled")
	_, rerr := os.ReadFile(filepath.Join(tmp, ".github", "release-notes", "v2.1.0.md"))
	assert.True(t, os.IsNotExist(rerr), "expected no release notes file to be written when canceled")
	commands := strings.Join(shell.Commands, "\n")
	assert.NotContains(t, commands, "git add", "expected no release notes file to be staged when canceled")
	assert.NotContains(t, commands, "git commit", "expected no commit to be made when canceled")
	assert.NotContains(t, out.Captured(), "git tag", "expected no tag command to be generated when canceled")
}

func TestReal_Release_SaveNotes_ReviewError(t *testing.T) {
	tmp := t.TempDir()
	shell := executor.NewMock()
	shell.Output = "https://github.com/volodya-lombrozo/aidy.git"
	mgit := git.NewMockWithDirAndShell(tmp, shell)
	out := output.NewMock()
	out.ReviewErr = fmt.Errorf("review failed")
	raidy := &real{git: mgit, ai: ai.NewMockAI(), reviewer: out, textreviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", true)

	assert.Error(t, err, "expected an error when reviewing release notes fails")
	assert.Contains(t, err.Error(), "failed to review release notes", "expected error message about reviewing release notes")
}

func TestReal_Release_SaveNotes_GitLab(t *testing.T) {
	tmp := t.TempDir()
	shell := executor.NewMock()
	shell.Output = "https://gitlab.com/volodya-lombrozo/aidy.git"
	mgit := git.NewMockWithDirAndShell(tmp, shell)
	out := output.NewMock()
	raidy := &real{git: mgit, ai: ai.NewMockAI(), reviewer: out, textreviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", true)

	require.NoError(t, err, "expected no error during release")
	notes, rerr := os.ReadFile(filepath.Join(tmp, ".gitlab", "release-notes", "v2.1.0.md"))
	require.NoError(t, rerr, "expected release notes file to be written under .gitlab")
	assert.Contains(t, string(notes), "Mock Release Notes", "expected release notes file to contain generated notes")
}

func TestReal_Release_SaveNotes_UnknownHost(t *testing.T) {
	tmp := t.TempDir()
	shell := executor.NewMock()
	shell.Output = "https://bitbucket.org/volodya-lombrozo/aidy.git"
	mgit := git.NewMockWithDirAndShell(tmp, shell)
	out := output.NewMock()
	raidy := &real{git: mgit, ai: ai.NewMockAI(), reviewer: out, textreviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", true)

	assert.Error(t, err, "expected error when the git host can't be determined")
	assert.Contains(t, err.Error(), "no known git host", "expected error to mention the unknown host")
}

func TestReal_Release_SkipsSavingNotesByDefault(t *testing.T) {
	tmp := t.TempDir()
	shell := executor.NewMock()
	shell.Output = "https://bitbucket.org/volodya-lombrozo/aidy.git"
	mgit := git.NewMockWithDirAndShell(tmp, shell)
	out := output.NewMock()
	raidy := &real{git: mgit, ai: ai.NewMockAI(), reviewer: out, textreviewer: out, logger: log.NewMock()}

	err := raidy.Release("minor", "origin", false)

	require.NoError(t, err, "expected no error when notes saving is disabled, even with an unknown host")
	_, rerr := os.ReadFile(filepath.Join(tmp, ".github", "release-notes", "v2.1.0.md"))
	assert.True(t, os.IsNotExist(rerr), "expected no release notes file to be written when --notes is not set")
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
		{"bug/PROJ-42", "PROJ-42"},
		{"feature/PROJ-17", "PROJ-17"},
		{"fix/ISSUE-100", "ISSUE-100"},
		{"chore/AB-99-cleanup", "AB-99"},
		{"feature/proj-42", "42"},
	}

	for _, test := range tests {
		result := inumber(test.branchName)
		if result != test.expected {
			t.Errorf("For branch name '%s', expected '%s', got '%s'", test.branchName, test.expected, result)
		}
	}
}

func TestBranchName(t *testing.T) {
	assert.Equal(t, "283-fix-mr-body-flag", branchName("283", "`fix-mr-body-flag`"))
	assert.Equal(t, "42-fix-bug", branchName("42", "fix-bug"))
}

func TestBranchName_PicksFirstMarkedSuggestionFromNoisyReply(t *testing.T) {
	reply := "Based on this issue, here are some suggested git branch names:\n\n" +
		"## Primary Suggestions:\n" +
		"1. **fix-branch-sanitize** - cleans the model output\n" +
		"2. **start-checkout** - fixes start\n\n" +
		"## Recommended Choice\n" +
		"Use `fix-branch-sanitize` because it is short."

	assert.Equal(t, "342-fix-branch-sanitize", branchName("342", reply))
}

func TestBranchName_UsesFirstNonEmptyLine(t *testing.T) {
	assert.Equal(t, "7-fix-login", branchName("7", "\n  Fix_Login \nexplanation follows"))
}

func TestBranchName_KeepsOnlyLowercaseLettersDigitsAndHyphens(t *testing.T) {
	assert.Equal(t, "5-feat-add-api-v2", branchName("5", "Feat/Add: API (v2)!"))
}

func TestBranchName_DropsDuplicatedIssueNumber(t *testing.T) {
	assert.Equal(t, "42-fix-bug", branchName("42", "42_fix_bug"))
}

func TestBranchName_CapsLength(t *testing.T) {
	name := branchName("42", strings.Repeat("word-", 20))

	assert.LessOrEqual(t, len(name), len("42-")+maxBranchSuffix)
	assert.False(t, strings.HasSuffix(name, "-"))
}

func TestBranchName_FallsBackToIssueNumber(t *testing.T) {
	assert.Equal(t, "342", branchName("342", ""))
	assert.Equal(t, "342", branchName("342", "  \n\n "))
	assert.Equal(t, "342", branchName("342", "### !!! ???"))
	assert.Equal(t, "342", branchName("342", "`342`"))
}

func TestReal_StartIssueChecksOutSanitizedBranch(t *testing.T) {
	shell := executor.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewMockAI(), github: github.NewMock(), cache: cache.NewMockAidyCache(), logger: log.Default()}

	err := raidy.StartIssue("42")

	require.NoError(t, err)
	assert.Equal(t, "git checkout -b 42-mock-branch-name", strings.TrimSpace(shell.Commands[len(shell.Commands)-1]))
}

func TestEscapeBackticks(t *testing.T) {
	input := "This is a `test` string with `backticks`."
	expected := "This is a \\`test\\` string with \\`backticks\\`."
	result := escapeBackticks(input)

	if result != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, result)
	}
}

func TestEscapeQuotes(t *testing.T) {
	input := `This is a "test" string with "quotes".`
	expected := `This is a \"test\" string with \"quotes\".`
	result := escapeQuotes(input)

	if result != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, result)
	}
}

func TestShellSafe(t *testing.T) {
	assert.Equal(t, `body with \"quotes\" inside`, shellSafe(`body with "quotes" inside`))
	assert.Equal(t, "top level quotes should be removed", shellSafe("`top level quotes should be removed`"))
}

func TestHealPRBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"keeps plain prose untouched",
			"Escapes double quotes in generated `aidy` commands.",
			"Escapes double quotes in generated `aidy` commands.",
		},
		{
			"removes a single header",
			"## Description\n\nEscapes double quotes in generated commands.",
			"Escapes double quotes in generated commands.",
		},
		{
			"removes all headers and keeps the prose",
			"## Description\n\nFix command failures.\n\n## Changes\n\nIntroduce `escapeQuotes()`.\n\n## Testing\n\nNew tests added.",
			"Fix command failures.\n\nIntroduce `escapeQuotes()`.\n\nNew tests added.",
		},
		{
			"removes headers of any level",
			"# One\n\nprose\n\n###### Six\n\nmore",
			"prose\n\nmore",
		},
		{
			"removes indented headers",
			"   ## Indented\n\nprose",
			"prose",
		},
		{
			"keeps issue references intact",
			"Fix the bug.\n\nFixes #238",
			"Fix the bug.\n\nFixes #238",
		},
		{
			"keeps a hash that is not a header",
			"#238 is fixed by this change.",
			"#238 is fixed by this change.",
		},
		{
			"trims surrounding whitespace",
			"\n\n  prose  \n\n",
			"prose",
		},
		{"handles an empty body", "", ""},
		{"handles a body made only of headers", "## Description\n\n## Changes\n", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, healPRBody(test.input))
		})
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

	alphanumericTests := []struct {
		actual   string
		expected string
	}{
		{"feat(#7523): Add feature", "feat(PROJ-17): Add feature"},
		{"feat(PROJ-99): Add feature", "feat(PROJ-17): Add feature"},
		{"fix(#7523): Fix bug", "fix(PROJ-17): Fix bug"},
		{"feat(#master): no change", "feat(#master): no change"},
	}
	for _, test := range alphanumericTests {
		result := healPRTitle(test.actual, "PROJ-17")
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

	logger := log.Default()

	assert.NotNil(t, logger, "Expected logger to be initialized")
	assert.IsType(t, &log.Short{}, logger, "Expected logger to be of type Short")
}

func TestInitLogger_SilentMode(t *testing.T) {
	InitLogger(true, false)

	logger := log.Default()

	assert.NotNil(t, logger, "Expected logger to be initialized")
	assert.IsType(t, &log.Silent{}, logger, "Expected logger to be of type Silent")
}

func TestInitLogger_DefaultMode(t *testing.T) {
	InitLogger(false, false)

	logger := log.Default()

	assert.NotNil(t, logger, "Expected logger to be initialized")
	assert.IsType(t, &log.Short{}, logger, "Expected logger to be of type Short")
}

func TestReal_Last_ReplaysSavedCommand(t *testing.T) {
	last := cache.NewMockLastCommand()
	last.WithLast("gh pr create --title \"edited\"")
	out := output.NewMock()
	raidy := &real{last: last, replayer: out}

	err := raidy.Last()

	require.NoError(t, err, "expected no error when repeating the last command")
	assert.Equal(t, "gh pr create --title \"edited\"", out.Last(), "expected the saved command to be replayed")
}

func TestReal_Last_FailsWithoutSavedCommand(t *testing.T) {
	out := output.NewMock()
	raidy := &real{last: cache.NewMockLastCommand(), replayer: out}

	err := raidy.Last()

	require.Error(t, err, "expected an error when there is nothing to repeat")
	assert.Contains(t, err.Error(), "no previous command")
	assert.Empty(t, out.Captured(), "expected nothing to be replayed")
}

func TestReal_StartIssueChecksOutJiraBranchByType(t *testing.T) {
	cases := map[string]string{
		"Bug":   "fix/PROJ-123",
		"Test":  "test/PROJ-123",
		"Story": "feat/PROJ-123",
		"Task":  "feat/PROJ-123",
	}
	for kind, expected := range cases {
		t.Run(kind, func(t *testing.T) {
			shell := executor.NewMock()
			raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewFailedMockAI(), github: github.NewMock(), jira: jira.NewMock(kind), cache: cache.NewMockAidyCache(), logger: log.NewMock()}

			err := raidy.StartIssue("PROJ-123")

			require.NoError(t, err)
			assert.Equal(t, []string{"git checkout -b " + expected}, trimmed(shell.Commands))
		})
	}
}

func TestReal_StartIssueAcceptsJiraLink(t *testing.T) {
	shell := executor.NewMock()
	gh := github.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewFailedMockAI(), github: gh, jira: jira.NewMock("Bug"), cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.StartIssue("https://company.atlassian.net/browse/PROJ-123")

	require.NoError(t, err)
	assert.Equal(t, []string{"git checkout -b fix/PROJ-123"}, trimmed(shell.Commands))
	assert.Zero(t, gh.Descriptions, "github should not be asked about jira issues")
}

func TestReal_StartIssueUsesFeatPrefixWithoutJiraConnection(t *testing.T) {
	shell := executor.NewMock()
	logger := log.NewMock()
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewFailedMockAI(), github: github.NewMock(), jira: jira.NewOffline(), cache: cache.NewMockAidyCache(), logger: logger}

	err := raidy.StartIssue("PROJ-123")

	require.NoError(t, err)
	assert.Equal(t, []string{"git checkout -b feat/PROJ-123"}, trimmed(shell.Commands))
	assert.Contains(t, strings.Join(logger.Messages, "\n"), "jira connection is not configured")
}

func TestReal_StartIssueUsesFeatPrefixWhenJiraFails(t *testing.T) {
	shell := executor.NewMock()
	broken := jira.NewMock("Bug")
	broken.Error = fmt.Errorf("unauthorized")
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewFailedMockAI(), github: github.NewMock(), jira: broken, cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.StartIssue("PROJ-123")

	require.NoError(t, err)
	assert.Equal(t, []string{"git checkout -b feat/PROJ-123"}, trimmed(shell.Commands))
}

func TestReal_StartIssueJiraCheckoutError(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("error checking out branch")
	raidy := &real{git: git.NewMockWithShell(shell), ai: ai.NewMockAI(), github: github.NewMock(), jira: jira.NewMock("Bug"), cache: cache.NewMockAidyCache(), logger: log.NewMock()}

	err := raidy.StartIssue("PROJ-123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error starting issue PROJ-123")
}

func TestNewJira_IsOfflineWithoutSettings(t *testing.T) {
	client, err := NewJira(config.NewMock())

	require.NoError(t, err)
	_, err = client.IssueType("PROJ-1")
	assert.ErrorContains(t, err, "jira connection is not configured")
}

func TestNewJira_FailsOnConfigError(t *testing.T) {
	conf := config.NewMock()
	conf.Error = fmt.Errorf("broken config")

	_, err := NewJira(conf)

	assert.ErrorContains(t, err, "broken config")
}

func trimmed(commands []string) []string {
	res := make([]string, 0, len(commands))
	for _, command := range commands {
		res = append(res, strings.TrimSpace(command))
	}
	return res
}
