package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/cache"
	"github.com/volodya-lombrozo/aidy/internal/git"
)

const JsonIssue = `{
  "title": "Title",
  "body": "Body"
}`

func TestRealGithub_Description(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(JsonIssue)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()
	gh := NewGithub(ts.URL, git.NewMock(), "", cache.NewMockAidyCache())

	description, err := gh.Description("123")

	require.NoError(t, err, "Description should not return an error")
	assert.Equal(t, fmt.Sprintf("Title: '%s'\nBody: '%s'", "Title", "Body"), description, "Description should match expected value")
}

func TestRealGithub_Description_NotNumber(t *testing.T) {
	gh := NewGithub("http://google.com", git.NewMock(), "", cache.NewMockAidyCache())
	issueNumber := "not-a-number"

	description, err := gh.Description(issueNumber)

	require.NoError(t, err, "Description should not return an error for non-numeric issue number")
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(json)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()
	gh := NewGithub(ts.URL, git.NewMock(), "", cache.NewMockAidyCache())

	labels, err := gh.Labels()

	require.NoError(t, err, "Labels should not return an error")
	expected := []string{"duplicate", "wontfix"}
	for i, label := range labels {
		assert.Equal(t, expected[i], label)
	}
}

func TestRealGithub_Remotes(t *testing.T) {
	gh := NewGithub("", git.NewMock(), "", cache.NewMockAidyCache())

	actual, err := gh.Remotes()

	require.NoError(t, err, "Remotes should not return an error")
	expected := []string{"volodya-lombrozo/aidy", "volodya-lombrozo/forked-aidy"}
	assert.Equal(t, expected, actual)
}
