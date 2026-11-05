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

func TestCodexCli_UsesReadOnlyEphemeralExecutionAndModel(t *testing.T) {
	cli := fakeCodexCli(t, "args", "gpt-test", false, "en")

	result, err := cli.SuggestBranch("Test description")

	require.NoError(t, err)
	assert.Contains(t, result, "exec --ephemeral --sandbox read-only --model gpt-test -")
}

func TestCodexCli_UsesDefaultModel(t *testing.T) {
	cli := fakeCodexCli(t, "args", "", false, "en")

	result, err := cli.SuggestBranch("Test description")

	require.NoError(t, err)
	assert.Contains(t, result, "--model "+codexCliDefaultModel)
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
