package ai

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const diff = `
diff --git a/ai/mockai.go b/ai/mockai.go
index 97c5ff0..2eea6a0 100644
--- a/ai/mockai.go
+++ b/ai/mockai.go
@@ -3 +3,4 @@ package ai
-import "fmt"
+import (
+       "fmt"
+       "strings"
+)
`

func TestMockGenerateCommitMessage(t *testing.T) {
	mockAI := NewMockAI()
	issue := "100"
	expected := fmt.Sprintf("feat(#%s): %s", issue, "changed files: ai/mockai.go")

	msg, err := mockAI.CommitMessage(issue, diff)

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, msg, "Expected commit message to match")
}

func TestMockGenerateLabels(t *testing.T) {
	mockAI := NewMockAI()
	labels := []string{"bug", "feature"}

	actual, err := mockAI.IssueLabels("issue", labels)

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, labels, actual)
}

func TestMock_GeneratePrTitle(t *testing.T) {
	mockAI := NewMockAI()
	branchName := "feature-branch"
	expected := fmt.Sprintf("mock title for '%s' with issue #issue and summary: summary", branchName)

	title, err := mockAI.PrTitle(branchName, "diff", "issue", "summary")

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, title, "Expected title to match")
}

func TestMock_GenerateIssueTitle(t *testing.T) {
	mockAI := NewMockAI()
	input := "issue input"
	expected := fmt.Sprintf("mock issue title for '%s' with summary: summary", input)

	title, err := mockAI.IssueTitle(input, "summary")

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, title, "Expected issue title to match")
}

func TestMock_GenerateIssueBody(t *testing.T) {
	mockAI := NewMockAI()
	input := "issue input"
	expected := fmt.Sprintf("mock issue body for '%s' with summary: summary", input)

	body, err := mockAI.IssueBody(input, "summary")

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, body, "Expected issue body to match")
}

func TestMock_GenerateBody(t *testing.T) {
	ai := NewMockAI()
	expected := "mock body for issue #issue and summary: summary\n\ndiff:\ndiff"
	body, err := ai.PrBody("diff", "issue", "summary")

	require.NoError(t, err, "Expected no error")
	assert.Contains(t, body, expected, "Expected body to match")
}

func TestMockAI_IssueLabels(t *testing.T) {
	mockAI := NewMockAI()
	labels := []string{"bug", "feature"}

	actual, err := mockAI.IssueLabels("issue", labels)

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, labels, actual, "Expected issue labels to match")
}

func TestMockAI_Summary(t *testing.T) {
	mockAI := NewMockAI()
	readme := "README content"
	expected := "summary: " + readme

	summary, err := mockAI.Summary(readme)

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, summary, "Expected summary to match")
}

func TestMockAI_SuggestBranch(t *testing.T) {
	mockAI := NewMockAI()
	expected := "mock-branch-name"

	branch, err := mockAI.SuggestBranch("description")

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, branch, "Expected branch name to match")
}

func TestMockAI_ReleaseNotes(t *testing.T) {
	mockAI := NewMockAI()
	changes := "Some changes"
	expected := fmt.Sprintf("Mock Release Notes\n\n%s", changes)

	notes, err := mockAI.ReleaseNotes(changes)

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, notes, "Expected release notes to match")
}

func TestFailedMockAI_ReleaseNotes(t *testing.T) {
	mockAI := NewFailedMockAI()
	changes := "Some changes"

	_, err := mockAI.ReleaseNotes(changes)

	require.Error(t, err, "Expected an error when generating release notes with failed mock")
	assert.Contains(t, err.Error(), "failed to generate release notes", "Expected error message to indicate failure")
}

func TestFailedMockAI_SuggestBranch(t *testing.T) {
	mockAI := NewFailedMockAI()
	description := "description"

	_, err := mockAI.SuggestBranch(description)

	require.Error(t, err, "Expected an error when suggesting branch with failed mock")
	assert.Contains(t, err.Error(), "failed to suggest branch", "Expected error message to indicate failure")
}
