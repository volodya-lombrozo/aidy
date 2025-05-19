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

	description := mockGithub.Description(issueNumber)
	assert.Equal(t, expectedDescription, description, "Description should match expected value")
}

func TestMockGithub_Labels(t *testing.T) {
	mockGithub := &MockGithub{}
	expected := []string{"bug", "documentation", "question"}
	labels := mockGithub.Labels()
	for i, label := range labels {
		assert.Equal(t, expected[i], label)
	}
}

func TestRealGithub_Description(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(JsonIssue)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()
	realGithub := NewGithub(ts.URL, git.NewMock(), "", cache.NewMockAidyCache())

	description := realGithub.Description("123")

	assert.Equal(t, fmt.Sprintf("Title: '%s'\nBody: '%s'", "Title", "Body"), description, "Description should match expected value")
}

func TestRealGithub_Description_NotNumber(t *testing.T) {
	realGithub := NewGithub("http://google.com", git.NewMock(), "", cache.NewMockAidyCache())
	issueNumber := "not-a-number"

	description := realGithub.Description(issueNumber)

	assert.Equal(t, "Invalid issue number: 'not-a-number'", description)
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
	realGithub := NewGithub(ts.URL, git.NewMock(), "", cache.NewMockAidyCache())

	labels := realGithub.Labels()

	expected := []string{"duplicate", "wontfix"}
	for i, label := range labels {
		assert.Equal(t, expected[i], label)
	}
}

func TestRealGithub_Remotes(t *testing.T) {
	gh := NewGithub("", git.NewMock(), "", cache.NewMockAidyCache())

	actual := gh.Remotes()

	expected := []string{"volodya-lombrozo/aidy", "volodya-lombrozo/forked-aidy"}
	assert.Equal(t, expected, actual)
}

func TestMockGithub_Remotes(t *testing.T) {
	expected := []string{"volodya-lombrozo/aidy", "volodya-lombrozo/jtcop"}
	gh := &MockGithub{}

	acutal := gh.Remotes()

	assert.Equal(t, expected, acutal)
}
