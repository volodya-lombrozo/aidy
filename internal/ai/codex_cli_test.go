package ai

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeCodexMode = "AIDY_FAKE_CODEX_MODE"

func fakeCodex(mode string) int {
	switch mode {
	case "stdin":
		input, _ := io.ReadAll(os.Stdin)
		fmt.Print(string(input))
	case "args":
		fmt.Println(strings.Join(os.Args[1:], " "))
	case "env":
		fmt.Printf("key=%s\n", os.Getenv("OPENAI_API_KEY"))
	case "stderr":
		fmt.Fprintln(os.Stderr, "warning")
		fmt.Println("answer")
	case "fail":
		fmt.Fprintln(os.Stderr, "not logged in")
		return 1
	case "labels":
		fmt.Println("bug")
	}
	return 0
}

func TestCodexCli_SendsPromptThroughStdin(t *testing.T) {
	cli := fakeCodexCli(t, "stdin", "", false, "en")

	result, err := cli.Summary("Test README content")

	require.NoError(t, err)
	assert.Contains(t, result, "helpful assistant writing project summaries")
	assert.Contains(t, result, "Test README content")
}

func TestCodexCli_UsesAidyPrompts(t *testing.T) {
	cli := fakeCodexCli(t, "stdin", "", false, "en")

	releaseNotes, err := cli.ReleaseNotes("changes")
	require.NoError(t, err)
	assert.Contains(t, releaseNotes, "generating GitHub release notes")

	prTitle, err := cli.PrTitle("12", "diff", "issue", "summary")
	require.NoError(t, err)
	assert.Contains(t, prTitle, "generating Git commit titles")

	issueTitle, err := cli.IssueTitle("input", "summary")
	require.NoError(t, err)
	assert.Contains(t, issueTitle, "creating GitHub issue titles")

	issueBody, err := cli.IssueBody("input", "summary")
	require.NoError(t, err)
	assert.Contains(t, issueBody, "writing GitHub issue descriptions")

	commit, err := cli.CommitMessage("12", "diff", "issue")
	require.NoError(t, err)
	assert.Contains(t, commit, "writing commit messages")
}

func TestCodexCli_UsesReadOnlyEphemeralExecutionAndModel(t *testing.T) {
	cli := fakeCodexCli(t, "args", "gpt-test", false, "en")

	result, err := cli.SuggestBranch("Test description")

	require.NoError(t, err)
	assert.Contains(t, result, "exec --ephemeral --ignore-user-config --ignore-rules --skip-git-repo-check --cd ")
	assert.Contains(t, result, "--sandbox read-only --model gpt-test --config model_reasoning_effort=low --config model_instructions_file=")
	assert.Contains(t, result, "--config web_search=\"disabled\" -")
	assert.Contains(t, result, "aidy-codex-")
}

func TestCodexCli_UsesDefaultModel(t *testing.T) {
	cli := fakeCodexCli(t, "args", "", false, "en")

	result, err := cli.SuggestBranch("Test description")

	require.NoError(t, err)
	assert.Contains(t, result, "--model "+codexCliDefaultModel)
}

func TestCodexCli_UsesModelSpecificDefaultReasoningEffort(t *testing.T) {
	for _, test := range []struct {
		model  string
		effort string
	}{
		{model: "gpt-6-luna", effort: "none"},
		{model: "gpt-test", effort: "low"},
	} {
		t.Run(test.model, func(t *testing.T) {
			cli := fakeCodexCli(t, "args", test.model, false, "en")

			result, err := cli.SuggestBranch("Test description")

			require.NoError(t, err)
			assert.Contains(t, result, "--config model_reasoning_effort="+test.effort)
		})
	}
}

func TestCodexCli_UsesConfiguredReasoningEffort(t *testing.T) {
	cli := NewCodexCliWithReasoningEffort("gpt-6-luna", "low", false, "en").(*CodexCli)
	cli.binary = func() string {
		self, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		return self
	}()
	t.Setenv(fakeCodexMode, "args")

	result, err := cli.SuggestBranch("Test description")

	require.NoError(t, err)
	assert.Contains(t, result, "--config model_reasoning_effort=low")
}

func TestCodexCli_UsesPriorityTierForLunaAndHonorsOverride(t *testing.T) {
	for _, test := range []struct {
		name string
		tier string
		want string
	}{
		{name: "Luna defaults to priority", want: "priority"},
		{name: "configured default disables priority", tier: "default", want: "default"},
	} {
		t.Run(test.name, func(t *testing.T) {
			cli := NewCodexCliWithSettings("gpt-6-luna", "low", test.tier, false, "en").(*CodexCli)
			cli.binary = func() string {
				self, err := os.Executable()
				require.NoError(t, err)
				return self
			}()
			t.Setenv(fakeCodexMode, "args")

			result, err := cli.SuggestBranch("Test description")

			require.NoError(t, err)
			assert.Contains(t, result, "--config service_tier="+test.want)
		})
	}
}

func TestCodexCli_HidesOpenAIAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "secret-key")
	cli := fakeCodexCli(t, "env", "", false, "en")

	result, err := cli.SuggestBranch("Test description")

	require.NoError(t, err)
	assert.Equal(t, "key=", result)
}

func TestCodexCli_HandlesStderrAndFailures(t *testing.T) {
	cli := fakeCodexCli(t, "stderr", "", false, "en")
	result, err := cli.SuggestBranch("Test description")
	require.NoError(t, err)
	assert.Equal(t, "answer", result)

	cli = fakeCodexCli(t, "fail", "", false, "en")
	_, err = cli.SuggestBranch("Test description")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not logged in")
}

func TestCodexCli_FailsOnEmptyOutput(t *testing.T) {
	cli := fakeCodexCli(t, "empty", "", false, "en")

	_, err := cli.SuggestBranch("Test description")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no content")
}

func TestCodexCli_AppendsSummaryAndLanguage(t *testing.T) {
	cli := fakeCodexCli(t, "stdin", "", true, "fr")

	result, err := cli.PrBody("Test diff", "Test issue", "Test summary")

	require.NoError(t, err)
	assert.Contains(t, result, "Test summary")
	assert.Contains(t, result, "respond entirely in fr language")
}

func TestCodexCli_IssueLabels(t *testing.T) {
	cli := fakeCodexCli(t, "labels", "", false, "en")

	labels, err := cli.IssueLabels("Test issue", []string{"bug", "enhancement"})

	require.NoError(t, err)
	assert.Equal(t, []string{"bug"}, labels)

	cli = fakeCodexCli(t, "fail", "", false, "en")
	_, err = cli.IssueLabels("Test issue", []string{"bug"})
	require.Error(t, err)
}

func TestCodexCli_ReportsMissingBinary(t *testing.T) {
	cli := NewCodexCli("", false, "en").(*CodexCli)
	cli.binary = "aidy-missing-codex-binary"

	_, err := cli.SuggestBranch("description")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in PATH")
}

func TestCodexCli_ReportsTemporaryDirectoryFailure(t *testing.T) {
	cli := fakeCodexCli(t, "stdin", "", false, "en")
	cli.tempDir = func() (string, error) { return "", fmt.Errorf("temporary directory unavailable") }

	_, err := cli.SuggestBranch("description")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "temporary directory unavailable")
}

func fakeCodexCli(t *testing.T, mode, model string, summary bool, language string) *CodexCli {
	t.Helper()
	t.Setenv(fakeCodexMode, mode)
	self, err := os.Executable()
	require.NoError(t, err)
	cli := NewCodexCli(model, summary, language).(*CodexCli)
	cli.binary = self
	return cli
}
