package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicAI_Summary(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL
	expected := "Test README content"

	result, err := ai.Summary(expected)

	require.NoError(t, err, "Expected no error when generating summary")
	assert.Contains(t, result, "generate a short, single-paragraph summary", "Echo server should return a command")
	assert.Contains(t, result, expected, "Expected summary to contain README content")
}

func TestAnthropicAI_ReleaseNotes(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", false, "en").(*Anthropic)
	ai.url = server.URL
	expected := "Test changes"

	result, err := ai.ReleaseNotes(expected)

	require.NoError(t, err, "Expected no error when generating release notes")
	assert.Contains(t, result, "Generate clear, concise release notes", "Echo server should return a command")
	assert.Contains(t, result, expected, "Expected release notes to contain changes")
}

func TestAnthropicAI_PrTitle(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL
	expectedDiff := "Test diff"
	expectedIssue := "Test issue"
	expectedNumber := "42"

	result, err := ai.PrTitle(expectedNumber, expectedDiff, expectedIssue, "")

	require.NoError(t, err, "Expected no error when generating PR title")
	assert.Contains(t, result, "generate a one-line PR title", "Echo server should return a command")
	assert.Contains(t, result, expectedDiff, "Expected PR title to contain diff")
	assert.Contains(t, result, expectedIssue, "Expected PR title to contain issue")
}

func TestAnthropicAI_PrBody(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL
	expectedDiff := "Test diff"
	expectedIssue := "Test issue"

	result, err := ai.PrBody(expectedDiff, expectedIssue, "")

	require.NoError(t, err, "Expected no error when generating PR body")
	assert.Contains(t, result, "generate a well-structured pull request body", "Echo server should return a command")
	assert.Contains(t, result, expectedDiff, "Expected PR body to contain diff")
	assert.Contains(t, result, expectedIssue, "Expected PR body to contain issue")
}

func TestAnthropicAI_IssueTitle(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL
	expectedInput := "Test input"

	result, err := ai.IssueTitle(expectedInput, "")

	require.NoError(t, err, "Expected no error when generating issue title")
	assert.Contains(t, result, "Generate a one-line issue title", "Echo server should return a command")
	assert.Contains(t, result, expectedInput, "Expected issue title to contain input")
}

func TestAnthropicAI_IssueBody(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL
	expectedInput := "Test input"

	result, err := ai.IssueBody(expectedInput, "")

	require.NoError(t, err, "Expected no error when generating issue body")
	assert.Contains(t, result, "Generate an issue body", "Echo server should return a command")
	assert.Contains(t, result, expectedInput, "Expected issue body to contain input")
}

func TestAnthropicAI_IssueLabels(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL
	expectedIssue := "Test issue"
	availableLabels := []string{"bug", "feature", "enhancement"}

	result, err := ai.IssueLabels(expectedIssue, availableLabels)

	require.NoError(t, err, "Expected no error when generating issue labels")
	for _, label := range availableLabels {
		assert.Contains(t, result, label, "Expected issue labels to contain available labels")
	}
}

func TestAnthropicAI_CommitMessage(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL
	expectedDiff := "Test diff"
	expectedNumber := "42"

	result, err := ai.CommitMessage(expectedNumber, expectedDiff, "")

	require.NoError(t, err, "Expected no error when generating commit message")
	assert.Contains(t, result, "generate a single-line commit message", "Echo server should return a command")
	assert.Contains(t, result, expectedDiff, "Expected commit message to contain diff")
}

func TestAnthropicAI_SuggestBranch(t *testing.T) {
	server := anthropicEchoServer(t)
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL
	expectedDescr := "Test description"

	result, err := ai.SuggestBranch(expectedDescr)

	require.NoError(t, err, "Expected no error when suggesting branch name")
	assert.Contains(t, result, "Generate a branch name", "Echo server should return a command")
	assert.Contains(t, result, expectedDescr, "Expected branch name to contain description")
}

func TestAnthropicAI_Handle404Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()
	ai := NewAnthropic("test-token", "", true, "en").(*Anthropic)
	ai.url = server.URL

	_, err := ai.Summary("Test README content")

	require.Error(t, err, "Expected an error when server returns 404")
	assert.Contains(t, err.Error(), "API error: Not Found", "Expected error message to contain 'API error: Not Found'")
}

func TestAnthropicAI_DefaultModel(t *testing.T) {
	ai := NewAnthropic("test-token", "", false, "en").(*Anthropic)
	assert.Equal(t, anthropicDefaultModel, ai.model, "Expected default model to be set")
}

func TestAnthropicAI_CustomModel(t *testing.T) {
	ai := NewAnthropic("test-token", "claude-opus-4-7", false, "en").(*Anthropic)
	assert.Equal(t, "claude-opus-4-7", ai.model, "Expected custom model to be set")
}

func anthropicEchoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err, "Failed to read request body")
		var request anthropicRequest
		err = json.Unmarshal(body, &request)
		require.NoError(t, err, "Failed to unmarshal request body")
		w.WriteHeader(http.StatusOK)
		message := ""
		if len(request.Messages) > 0 {
			message = request.Messages[0].Content
		}
		message = replaceChars(message)
		resp := fmt.Sprintf(`{"content":[{"type":"text","text":"%s"}]}`, message)
		_, err = w.Write([]byte(resp))
		require.NoError(t, err, "Failed to write response")
	}))
}

func replaceChars(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\n':
			result = append(result, '\\', 'n')
		case '"':
			result = append(result, '\'')
		default:
			result = append(result, s[i])
		}
	}
	return string(result)
}
