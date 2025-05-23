package aidy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockAidy_Release(t *testing.T) {
	aidy := NewMock()
	err := aidy.Release("daily", "repo-name")
	assert.NoError(t, err)
	assert.Contains(t, aidy.Logs(), "Release called with interval: daily, repo: repo-name")
}

func TestMockAidy_PrintConfig(t *testing.T) {
	aidy := NewMock()
	aidy.PrintConfig()
	assert.Contains(t, aidy.Logs(), "PrintConfig called")
}

func TestMockAidy_Commit(t *testing.T) {
	aidy := NewMock()
	aidy.Commit()
	assert.Contains(t, aidy.Logs(), "Commit called")
}

func TestMockAidy_Squash(t *testing.T) {
	aidy := NewMock()
	aidy.Squash()
	assert.Contains(t, aidy.Logs(), "Squash called")
}

func TestMockAidy_PullRequest(t *testing.T) {
	aidy := NewMock()
	aidy.PullRequest()
	assert.Contains(t, aidy.Logs(), "PullRequest called")
}

func TestMockAidy_Issue(t *testing.T) {
	aidy := NewMock()
	aidy.Issue("test-task")
	assert.Contains(t, aidy.Logs(), "Issue called with task: test-task")
}

func TestMockAidy_Heal(t *testing.T) {
	aidy := NewMock()
	aidy.Heal()
	assert.Contains(t, aidy.Logs(), "Heal called")
}

func TestMockAidy_Append(t *testing.T) {
	aidy := NewMock()
	aidy.Append()
	assert.Contains(t, aidy.Logs(), "Append called")
}

func TestMockAidy_Clean(t *testing.T) {
	aidy := NewMock()
	aidy.Clean()
	assert.Contains(t, aidy.Logs(), "Clean called")
}

func TestMockAidy_Diff(t *testing.T) {
	aidy := NewMock()
	aidy.Diff()
	assert.Contains(t, aidy.Logs(), "Diff called")
}

func TestMockAidy_StartIssue(t *testing.T) {
	aidy := NewMock()
	err := aidy.StartIssue("123")
	assert.NoError(t, err)
	assert.Contains(t, aidy.Logs(), "StartIssue called with number: 123")
}
