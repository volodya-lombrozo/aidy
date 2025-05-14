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
	mockAI := &MockAI{}
	issue := "100"
	expected := fmt.Sprintf("feat(#%s): %s", issue, "changed files: ai/mockai.go")

	msg, err := mockAI.CommitMessage(issue, diff)

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, msg, "Expected commit message to match")
}

func TestMockGenerateLabels(t *testing.T) {
	mockAI := &MockAI{}
	labels := []string{"bug", "feature"}

	actual, err := mockAI.IssueLabels("issue", labels)

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, labels, actual)
}

func TestMockGenerateTitle(t *testing.T) {
	mockAI := &MockAI{}
	branchName := "feature-branch"
	expected := "'Mock Title for " + branchName + "'"

	title, err := mockAI.PrTitle(branchName, "diff", "issue", "summary")

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, title, "Expected title to match")
}

func TestMockGenerateIssueTitle(t *testing.T) {
	mockAI := &MockAI{}
	userInput := "issue input"
	expected := "'Mock Issue Title for " + userInput + "'"

	title, err := mockAI.IssueTitle(userInput, "summary")

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, title, "Expected issue title to match")
}

func TestMockGenerateIssueBody(t *testing.T) {
	mockAI := &MockAI{}
	userInput := "issue input"
	expected := "Mock Issue Body for " + userInput

	body, err := mockAI.IssueBody(userInput, "summary")

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, body, "Expected issue body to match")
}

func TestMockGenerateBody(t *testing.T) {
	mockAI := &MockAI{}
	branchName := "feature-branch"
	expected := "Mock Body for " + branchName

	body, err := mockAI.PrBody(branchName, "diff", "issue", "summary")

	require.NoError(t, err, "Expected no error")
	assert.Equal(t, expected, body, "Expected body to match")
}
