package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volodya-lombrozo/aidy/cache"
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

func TestMockGithub_Labels(t *testing.T) {
	mockGithub := &MockGithub{}
	expected := []string{"bug", "documentation", "question"}
	labels := mockGithub.Labels()
	for i, label := range labels {
		assert.Equal(t, expected[i], label)
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

	realGithub := NewRealGithub(ts.URL, &git.MockGit{}, "", cache.NewMockCache())

	issueNumber := "123"
	description := realGithub.IssueDescription(issueNumber)
	expectedDescription := fmt.Sprintf("Title: '%s'\nBody: '%s'", "Title", "Body")

	if description != expectedDescription {
		t.Errorf("expected '%s', got '%s'", expectedDescription, description)
	}
}

func TestRealGithub_Labels(t *testing.T) {
	json := `
[
  {
    "id": 4737785601,
    "node_id": "LA_kwDOIVTLe88AAAABGmTfAQ",
    "url": "https://api.github.com/repos/volodya-lombrozo/jtcop/labels/duplicate",
    "name": "duplicate",
    "color": "cfd3d7",
    "default": true,
    "description": "This issue or pull request already exists"
  },
  {
    "id": 4737785618,
    "node_id": "LA_kwDOIVTLe88AAAABGmTfEg",
    "url": "https://api.github.com/repos/volodya-lombrozo/jtcop/labels/wontfix",
    "name": "wontfix",
    "color": "ffffff",
    "default": true,
    "description": "This will not be worked on"
  }
]`
	// Create a test server to mock GitHub API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(json)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()
	realGithub := NewRealGithub(ts.URL, &git.MockGit{}, "", cache.NewMockCache())

	labels := realGithub.Labels()

	expected := []string{"duplicate", "wontfix"}
	for i, label := range labels {
		assert.Equal(t, expected[i], label)
	}
}
