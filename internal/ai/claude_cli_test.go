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

const fakeClaudeMode = "AIDY_FAKE_CLAUDE_MODE"

func TestMain(m *testing.M) {
	if mode, ok := os.LookupEnv(fakeClaudeMode); ok {
		os.Exit(fakeClaude(mode))
	}
	os.Exit(m.Run())
}

func fakeClaude(mode string) int {
	switch mode {
	case "stdin":
		in, _ := io.ReadAll(os.Stdin)
		fmt.Print(string(in))
	case "args":
		fmt.Println(strings.Join(os.Args[1:], " "))
	case "env":
		fmt.Printf("key=%s\n", os.Getenv("ANTHROPIC_API_KEY"))
	case "stderr":
		fmt.Fprintln(os.Stderr, "warning")
		fmt.Println("answer")
	case "fail":
		fmt.Fprintln(os.Stderr, "not logged in")
		return 1
	case "bug":
		fmt.Println("bug")
	}
	return 0
}

func TestClaudeCli_SendsPromptThroughStdin(t *testing.T) {
	ai := fakeClaudeCli(t, "stdin", "", false, "en")

	result, err := ai.Summary("Test README content")

	require.NoError(t, err, "Expected no error when generating summary")
	assert.Contains(t, result, "generate a short, single-paragraph summary", "Fake cli should echo the prompt")
	assert.Contains(t, result, "Test README content", "Expected summary to contain README content")
}

func TestClaudeCli_PassesSystemPromptAndModel(t *testing.T) {
	ai := fakeClaudeCli(t, "args", "haiku", false, "en")

	result, err := ai.SuggestBranch("Test description")

	require.NoError(t, err, "Expected no error when suggesting branch")
	assert.Contains(t, result, "-p", "Expected headless mode")
	assert.Contains(t, result, "--no-session-persistence", "Expected no session transcripts on disk")
	assert.Contains(t, result, "--strict-mcp-config", "Expected MCP servers to be skipped")
	assert.Contains(t, result, "--setting-sources  ", "Expected user and project settings to be skipped")
	assert.Contains(t, result, "--disable-slash-commands", "Expected skills to be skipped")
	assert.Contains(t, result, "--system-prompt You are a helpful assistant suggesting branch names.", "Expected system prompt")
	assert.Contains(t, result, "--model haiku", "Expected configured model")
	assert.Contains(t, result, "--output-format text", "Expected text output")
}

func TestClaudeCli_UsesDefaultModel(t *testing.T) {
	ai := fakeClaudeCli(t, "args", "", false, "en")

	result, err := ai.SuggestBranch("Test description")

	require.NoError(t, err, "Expected no error when suggesting branch")
	assert.Contains(t, result, "--model sonnet", "Expected default model")
}

func TestClaudeCli_HidesAnthropicApiKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "secret-key")
	ai := fakeClaudeCli(t, "env", "", false, "en")

	result, err := ai.SuggestBranch("Test description")

	require.NoError(t, err, "Expected no error when suggesting branch")
	assert.Equal(t, "key=", result, "Expected ANTHROPIC_API_KEY to be removed from cli environment")
}

func TestClaudeCli_IgnoresStderr(t *testing.T) {
	ai := fakeClaudeCli(t, "stderr", "", false, "en")

	result, err := ai.SuggestBranch("Test description")

	require.NoError(t, err, "Expected no error when cli writes to stderr")
	assert.Equal(t, "answer", result, "Expected only stdout in result")
}

func TestClaudeCli_ReportsStderrOnFailure(t *testing.T) {
	ai := fakeClaudeCli(t, "fail", "", false, "en")

	_, err := ai.SuggestBranch("Test description")

	require.Error(t, err, "Expected error when cli fails")
	assert.Contains(t, err.Error(), "not logged in", "Expected stderr in error message")
}

func TestClaudeCli_FailsOnEmptyOutput(t *testing.T) {
	ai := fakeClaudeCli(t, "empty", "", false, "en")

	_, err := ai.SuggestBranch("Test description")

	require.Error(t, err, "Expected error when cli returns nothing")
}

func TestClaudeCli_FailsWhenBinaryIsMissing(t *testing.T) {
	ai := NewClaudeCli("", false, "en").(*ClaudeCli)
	ai.binary = "aidy-missing-claude-binary"

	_, err := ai.SuggestBranch("Test description")

	require.Error(t, err, "Expected error when binary is missing")
	assert.Contains(t, err.Error(), "not found in PATH", "Expected clear message about missing binary")
}

func TestClaudeCli_AppendsSummaryAndLanguage(t *testing.T) {
	ai := fakeClaudeCli(t, "stdin", "", true, "fr")

	result, err := ai.PrBody("Test diff", "Test issue", "Test summary")

	require.NoError(t, err, "Expected no error when generating PR body")
	assert.Contains(t, result, "Test summary", "Expected project summary in prompt")
	assert.Contains(t, result, "respond entirely in fr language", "Expected language instruction in prompt")
}

func TestClaudeCli_IssueLabels(t *testing.T) {
	ai := fakeClaudeCli(t, "bug", "", false, "en")

	labels, err := ai.IssueLabels("Test issue", []string{"bug", "enhancement"})

	require.NoError(t, err, "Expected no error when generating labels")
	assert.Equal(t, []string{"bug"}, labels, "Expected only labels mentioned in response")
}

func fakeClaudeCli(t *testing.T, mode, model string, summary bool, language string) *ClaudeCli {
	t.Helper()
	t.Setenv(fakeClaudeMode, mode)
	self, err := os.Executable()
	require.NoError(t, err, "Failed to locate test binary acting as fake claude cli")
	ai := NewClaudeCli(model, summary, language).(*ClaudeCli)
	ai.binary = self
	return ai
}
