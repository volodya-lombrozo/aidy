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

func (m *MockAI) ReleaseNotes(changes string) (string, error) {
	if m.fail {
		return "", fmt.Errorf("failed to generate release notes")
	}
	return fmt.Sprintf("Mock Release Notes\n\n%s", changes), nil
}

func (m *MockAI) PrTitle(branchName string, diff string, issue string, summary string) (string, error) {
	return fmt.Sprintf("mock title for '%s' with issue #%s and summary: %s", branchName, issue, summary), nil
}

func (m *MockAI) PrBody(branch string, diff string, issue string, summary string) (string, error) {
	return fmt.Sprintf("mock body for %s with issue #%s and summary: %s\n\ndiff:\n%s", branch, issue, summary, diff), nil
}

func (m *MockAI) IssueTitle(input string, summary string) (string, error) {
	return fmt.Sprintf("mock issue title for '%s' with summary: %s", input, summary), nil
}

func (m *MockAI) IssueBody(input string, summary string) (string, error) {
	return fmt.Sprintf("mock issue body for '%s' with summary: %s", input, summary), nil
}

func (m *MockAI) CommitMessage(issue string, diff string) (string, error) {
	return fmt.Sprintf("feat(#%s): %s", issue, summary(diff)), nil
}

func (m *MockAI) IssueLabels(issue string, available []string) ([]string, error) {
	return available, nil
}

func (m *MockAI) Summary(readme string) (string, error) {
	if m.fail {
		return "", fmt.Errorf("failed to generate summary")
	}
	return "summary: " + readme, nil
}

func (m *MockAI) SuggestBranch(descr string) (string, error) {
	if m.fail {
		return "", fmt.Errorf("failed to suggest branch")
	}
	return "mock-branch-name", nil
}

/*
Parse unified diff into a short summary string

	Input example:

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

	This function just looks for files that were changed
	it also ignores /dev/null file and prints only unique files
	Also it should remove a/ and b/ prefixes
*/
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
	if len(files) == 0 {
		return "no files changed"
	}
	return fmt.Sprintf("changed files: %s", strings.Join(files, ", "))
}
