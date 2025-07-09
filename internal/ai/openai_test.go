package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type echo struct {
}

func NewEcho() openClient {
	return &echo{}
}

func (m echo) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	if len(req.Messages) == 0 {
		return openai.ChatCompletionResponse{}, fmt.Errorf("no messages in request")
	}
	msg := req.Messages[0].Content
	if strings.Contains(msg, "successful") {
		return openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: msg}},
			},
		}, nil
	}
	if strings.Contains(msg, "error") {
		return openai.ChatCompletionResponse{}, fmt.Errorf("error during openai request")
	}
	return openai.ChatCompletionResponse{}, fmt.Errorf("nothing to answer")
}

func TestOpenAi_GeneratesTitle(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)

	title, err := openAI.PrTitle("123", "test diff", "successful issue-description", "project-summary")

	require.NoError(t, err, "Expected no error when generating PR title")
	assert.Contains(t, title, "generate a one-line PR title", "anser should contain the command to generate a PR title")
}

func TestOpenAi_GeneratesTitleWithError(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)

	_, err := openAI.PrTitle("123", "test diff", "error issue-description", "project-summary")

	require.Error(t, err, "Expected error when generating PR title with error input")
	assert.Equal(t, "error during openai request", err.Error(), "Expected error message to match mock response")
}

func TestOpenAi_GeneratesBody(t *testing.T) {
	openai := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)

	body, err := openai.PrBody("test diff", "successful issue-description", "project-summary")

	require.NoError(t, err, "Expected no error when generating PR body")
	assert.Contains(t, body, "generate a well-structured pull request body", "Expected PR body to match mock response")
}

func TestOpenAI_SuggestBranch(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)
	description := "successful description"

	branch, err := openAI.SuggestBranch(description)

	require.NoError(t, err, "Expected no error when suggesting branch name")
	assert.Contains(t, branch, description, "Expected branch name to contain description")
	assert.Contains(t, branch, "Generate a branch name", "Expected branch name to contain 'branch' keyword")
}

func TestOpenAI_SuggestBranch_Error(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)
	description := "error description"

	_, err := openAI.SuggestBranch(description)

	require.Error(t, err, "Expected error when suggesting branch with error input")
	assert.Equal(t, "error during openai request", err.Error(), "Expected error message to match mock response")
}

func TestOpenAI_ReleaseNotes(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)
	changes := "successful changes"

	notes, err := openAI.ReleaseNotes(changes)

	require.NoError(t, err, "Expected no error when generating release notes")
	assert.Contains(t, notes, changes, "Expected release notes to contain changes")
	assert.Contains(t, notes, "Generate clear, concise release notes", "Expected release notes to contain 'release notes' keyword")
}

func TestOpenAI_ReleaseNotes_Error(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)
	changes := "error changes"

	_, err := openAI.ReleaseNotes(changes)

	require.Error(t, err, "Expected error when generating release notes with error input")
	assert.Equal(t, "error during openai request", err.Error(), "Expected error message to match mock response")
}

func TestOpenAI_IssueTitle(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)
	input := "successful issue input"

	title, err := openAI.IssueTitle(input, "project-summary")

	require.NoError(t, err, "Expected no error when generating issue title")
	assert.Contains(t, title, input, "Expected issue title to contain input")
	assert.Contains(t, title, "Generate a one-line issue title", "Expected issue title to contain 'issue title' keyword")
}

func TestOpenAI_IssueBody(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)
	input := "successful issue input"

	body, err := openAI.IssueBody(input, "project-summary")

	require.NoError(t, err, "Expected no error when generating issue body")
	assert.Contains(t, body, input, "Expected issue body to contain input")
	assert.Contains(t, body, "Generate an issue body", "Expected issue body to contain 'issue description' keyword")
}

func TestOpenAI_CommitMessage(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)
	number := "123"
	diff := "successful diff"

	message, err := openAI.CommitMessage(number, diff, "")

	require.NoError(t, err, "Expected no error when generating commit message")
	assert.Contains(t, message, diff, "Expected commit message to contain diff")
	assert.Contains(t, message, "generate a single-line commit message", "Expected commit message to contain 'commit message' keyword")
}

func TestOpenAI_IssueLabels(t *testing.T) {
	openAI := NewOpenAIWithClient(NewEcho(), "test-model", 0.5, false)
	issue := "I successfully suggest using 'feature' label here"
	available := []string{"bug", "feature", "enhancement"}

	labels, err := openAI.IssueLabels(issue, available)

	require.NoError(t, err, "Expected no error when generating issue labels")
	assert.ElementsMatch(t, labels, available, "Expected issue labels to match available labels")
}
