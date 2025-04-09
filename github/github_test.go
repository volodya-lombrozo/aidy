package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/volodya-lombrozo/aidy/git"
)

const JsonIssue = `{
  "title": "Title",
  "body": "Body"
}`

func TestMockGithub_IssueDescription(t *testing.T) {
	mockGithub := &MockGithub{}
	issueNumber := "123"
	expectedDescription := "Mock description for issue #" + issueNumber

	description := mockGithub.IssueDescription(issueNumber)
	if description != expectedDescription {
		t.Errorf("expected %s, got %s", expectedDescription, description)
	}
}

func TestRealGithub_IssueDescription(t *testing.T) {
	// Create a test server to mock GitHub API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(JsonIssue)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()

	realGithub := NewRealGithub(ts.URL, &git.MockGit{}, "")

	issueNumber := "123"
	description := realGithub.IssueDescription(issueNumber)
	expectedDescription := fmt.Sprintf("Title: '%s'\nBody: '%s'", "Title", "Body")

	if description != expectedDescription {
		t.Errorf("expected '%s', got '%s'", expectedDescription, description)
	}
}
