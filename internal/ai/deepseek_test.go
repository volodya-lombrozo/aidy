package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekAI_Summary(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
	ai.url = server.URL
	expected := "Test README content"

	result, err := ai.Summary(expected)

	require.NoError(t, err, "Expected no error when generating summary")
	assert.Contains(t, result, "generate a short, single-paragraph summary", "Echo server should return a command")
	assert.Contains(t, result, expected, "Expected summary to contain README content")
}

func TestDeepSeekAI_ReleaseNotes(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", false).(*DeepSeek)
	ai.url = server.URL
	expected := "Test changes"

	result, err := ai.ReleaseNotes(expected)

	require.NoError(t, err, "Expected no error when generating release notes")
	assert.Contains(t, result, "Generate clear, concise release notes", "Echo server should return a command")
	assert.Contains(t, result, expected, "Expected release notes to contain changes")
}

func TestDeepSeekAI_PrTitle(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
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

func TestDeepSeekAI_PrBody(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
	ai.url = server.URL
	expectedDiff := "Test diff"
	expectedIssue := "Test issue"
	expectedNumber := "42"

	result, err := ai.PrBody(expectedNumber, expectedDiff, expectedIssue, "")

	require.NoError(t, err, "Expected no error when generating PR body")
	assert.Contains(t, result, "generate a well-structured pull request body", "Echo server should return a command")
	assert.Contains(t, result, expectedDiff, "Expected PR body to contain diff")
	assert.Contains(t, result, expectedIssue, "Expected PR body to contain issue")
}

func TestDeepSeekAI_IssueTitle(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
	ai.url = server.URL
	expectedInput := "Test input"

	result, err := ai.IssueTitle(expectedInput, "")

	require.NoError(t, err, "Expected no error when generating issue title")
	assert.Contains(t, result, "Generate a one-line issue title", "Echo server should return a command")
	assert.Contains(t, result, expectedInput, "Expected issue title to contain input")
}

func TestDeepSeekAI_IssueBody(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
	ai.url = server.URL
	expectedInput := "Test input"

	result, err := ai.IssueBody(expectedInput, "")

	require.NoError(t, err, "Expected no error when generating issue body")
	assert.Contains(t, result, "Generate an issue body", "Echo server should return a command")
	assert.Contains(t, result, expectedInput, "Expected issue body to contain input")
}

func TestDeepSeekAI_IssueLabels(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
	ai.url = server.URL
	expectedIssue := "Test issue"
	availableLabels := []string{"bug", "feature", "enhancement"}

	result, err := ai.IssueLabels(expectedIssue, availableLabels)

	require.NoError(t, err, "Expected no error when generating issue labels")
	for _, label := range availableLabels {
		assert.Contains(t, result, label, "Expected issue labels to contain available labels")
	}
}

func TestDeepSeekAI_CommitMessage(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
	ai.url = server.URL
	expectedDiff := "Test diff"
	expectedNumber := "42"

	result, err := ai.CommitMessage(expectedNumber, expectedDiff)

	require.NoError(t, err, "Expected no error when generating commit message")
	assert.Contains(t, result, "generate a single-line commit message", "Echo server should return a command")
	assert.Contains(t, result, expectedDiff, "Expected commit message to contain diff")
}

func TestDeepSeekAI_SuggestBranch(t *testing.T) {
	server := echoServer(t)
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
	ai.url = server.URL
	expectedDescr := "Test description"

	result, err := ai.SuggestBranch(expectedDescr)

	require.NoError(t, err, "Expected no error when suggesting branch name")
	assert.Contains(t, result, "Generate a branch name", "Echo server should return a command")
	assert.Contains(t, result, expectedDescr, "Expected branch name to contain description")
}

func TestDeepSeekAI_Handle404Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()
	ai := NewDeepSeek("test-token", true).(*DeepSeek)
	ai.url = server.URL

	_, err := ai.Summary("Test README content")

	require.Error(t, err, "Expected an error when server returns 404")
	assert.Contains(t, err.Error(), "API error: Not Found", "Expected error message to contain 'API error: Not Found'")
}

func echoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err, "Failed to read request body")
		var request chatRequest
		err = json.Unmarshal(body, &request)
		require.NoError(t, err, "Failed to unmarshal request body")
		require.NoError(t, err, "Failed to read request body")
		w.WriteHeader(http.StatusOK)
		message := strings.ReplaceAll(request.Messages[1].Content, "\n", "\\n")
		message = strings.ReplaceAll(message, "\"", "'")
		resp := fmt.Sprintf(`{"choices":[{"message":{"content":"%s"}}]}`, message)
		_, err = w.Write([]byte(resp))
		require.NoError(t, err, "Failed to write response")
	}))
}
