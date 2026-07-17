package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/executor"
)

func TestReal_MergeRequestByBranch(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = `[{"title": "MR Title", "description": "MR Body"}]`
	gl := NewGitlab(shell)

	title, body, err := gl.MergeRequestByBranch("feature-branch")

	require.NoError(t, err, "MergeRequestByBranch should not return an error")
	assert.Equal(t, "MR Title", title)
	assert.Equal(t, "MR Body", body)
	assert.Contains(t, shell.Commands[0], "glab mr list --source-branch feature-branch --output json")
}

func TestReal_MergeRequestByBranch_NotFound(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = `[]`
	gl := NewGitlab(shell)

	title, body, err := gl.MergeRequestByBranch("feature-branch")

	require.Error(t, err, "MergeRequestByBranch should return an error when no MR is found")
	assert.Contains(t, err.Error(), "no open merge request found for branch 'feature-branch'")
	assert.Empty(t, title)
	assert.Empty(t, body)
}

func TestReal_MergeRequestByBranch_Multiple(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = `[{"title": "First MR", "description": "First Body"}, {"title": "Second MR", "description": "Second Body"}]`
	gl := NewGitlab(shell)

	title, body, err := gl.MergeRequestByBranch("feature-branch")

	require.NoError(t, err, "MergeRequestByBranch should not return an error when multiple MRs are found")
	assert.Equal(t, "First MR", title)
	assert.Equal(t, "First Body", body)
}

func TestReal_MergeRequestByBranch_CommandError(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = assert.AnError
	gl := NewGitlab(shell)

	title, body, err := gl.MergeRequestByBranch("feature-branch")

	require.Error(t, err, "MergeRequestByBranch should return an error when the command fails")
	assert.Contains(t, err.Error(), "error fetching merge requests for branch 'feature-branch'")
	assert.Empty(t, title)
	assert.Empty(t, body)
}

func TestReal_MergeRequestByBranch_InvalidJson(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "not-json"
	gl := NewGitlab(shell)

	title, body, err := gl.MergeRequestByBranch("feature-branch")

	require.Error(t, err, "MergeRequestByBranch should return an error when the json is invalid")
	assert.Contains(t, err.Error(), "error parsing merge request json for branch 'feature-branch'")
	assert.Empty(t, title)
	assert.Empty(t, body)
}
