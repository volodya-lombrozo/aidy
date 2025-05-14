package ai

import (
	"fmt"
	"strings"
)

type MockAI struct {
	fail bool
}

func NewMockAI() AI {
	return &MockAI{fail: false}
}

func NewFailedMockAI() AI {
	return &MockAI{fail: true}
}

func (m *MockAI) PrTitle(branchName string, diff string, issue string, summary string) (string, error) {
	return "'Mock Title for " + branchName + "'", nil
}

func (m *MockAI) PrBody(branchName string, diff string, issue string, summary string) (string, error) {
	return "Mock Body for " + branchName, nil
}

func (m *MockAI) IssueTitle(userInput string, summary string) (string, error) {
	return "'Mock Issue Title for " + userInput + "'", nil
}

func (m *MockAI) IssueBody(userInput string, summary string) (string, error) {
	return "Mock Issue Body for " + userInput, nil
}

func (m *MockAI) CommitMessage(issue string, diff string) (string, error) {
	return fmt.Sprintf("feat(#%s): %s", issue, summary(diff)), nil
}

func (m *MockAI) IssueLabels(issue string, available []string) ([]string, error) {
	return available, nil
}

func (m *MockAI) Summary(readme string) (string, error) {
	return "summary: " + readme, nil
}

func (m *MockAI) SuggestBranch(descr string) (string, error) {
	if m.fail {
		return "", fmt.Errorf("failed to suggest branch")
	}
	return "mock-branch-name", nil
}

// Parse unified diff into a short summary string
// Input example:
//
// diff --git a/ai/mockai.go b/ai/mockai.go
// index 97c5ff0..2eea6a0 100644
// --- a/ai/mockai.go
// +++ b/ai/mockai.go
// @@ -3 +3,4 @@ package ai
// -import "fmt"
// +import (
// +       "fmt"
// +       "strings"
// +)
//
// This function just looks for files that were changed
// it also ignores /dev/null file and prints only unique files
// Also it should remove a/ and b/ prefixes
func summary(diff string) string {
	lines := strings.Split(strings.TrimSpace(diff), "\n")
	unique := make(map[string]struct{})
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}
			fileA := strings.TrimPrefix(parts[2], "a/")
			fileB := strings.TrimPrefix(parts[3], "b/")
			if fileA != "/dev/null" {
				unique[fileA] = struct{}{}
			}
			if fileB != "/dev/null" {
				unique[fileB] = struct{}{}
			}
		}
	}
	var files []string
	for file := range unique {
		files = append(files, file)
	}
	return fmt.Sprintf("changed files: %s", strings.Join(files, ", "))
}
