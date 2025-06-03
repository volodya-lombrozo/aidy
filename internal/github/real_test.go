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
	assert.Equal(t, "invalid issue number: 'not-a-number'", description)
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

func TestRealGithub_Description_NoRemote(t *testing.T) {
	gh := NewGithub("http://example.com", git.NewMock(), "", cache.NewMockAidyCache())
	gh.ch.WithRemote("")

	description, err := gh.Description("123")

	require.Error(t, err, "Description should return an error when no remote is set")
	assert.Equal(t, "cannot find a target repository to search for issue '123'", err.Error())
	assert.Empty(t, description, "Description should be empty on error")
}

func TestRealGithub_Labels_NoRemote(t *testing.T) {
	gh := NewGithub("http://example.com", git.NewMock(), "", cache.NewMockAidyCache())
	gh.ch.WithRemote("")

	labels, err := gh.Labels()

	require.Error(t, err, "Labels should return an error when request creation fails")
	assert.Contains(t, err.Error(), "cannot determine where to get labels, please set the target repository")
	assert.Nil(t, labels, "Labels should be nil on error")
}

func TestRealGithub_Labels_UrlParsingError(t *testing.T) {
	gh := NewGithub("\\invalid://url", git.NewMock(), "", cache.NewMockAidyCache())
	gh.ch.WithRemote("invalid-url")

	labels, err := gh.Labels()

	require.Error(t, err, "Labels should return an error when request creation fails")
	assert.Contains(t, err.Error(), "cannot create a new GET request to retrieve labels")
	assert.Nil(t, labels, "Labels should be nil on error")
}

func TestRealGithub_Labels_InvalidProtocol(t *testing.T) {
	gh := NewGithub("invalid://protocol", git.NewMock(), "", cache.NewMockAidyCache())
	gh.ch.WithRemote("invalid-protocol")

	labels, err := gh.Labels()

	require.Error(t, err, "Labels should return an error when request creation fails")
	assert.Contains(t, err.Error(), "error fetching issue labels")
	assert.Contains(t, err.Error(), "unsupported protocol scheme \"invalid\"")
	assert.Nil(t, labels, "Labels should be nil on error")
}

func TestRealGithub_Labels_DoRequestError(t *testing.T) {
	gh := NewGithub("http://example.com", git.NewMock(), "", cache.NewMockAidyCache())
	gh.ch.WithRemote("valid-repo")
	gh.client = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("simulated request error")
		}),
	}

	labels, err := gh.Labels()

	require.Error(t, err, "Labels should return an error when HTTP request fails")
	assert.Contains(t, err.Error(), "error fetching issue labels")
	assert.Contains(t, err.Error(), "simulated request error")
	assert.Nil(t, labels, "Labels should be nil on error")
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRealGithub_Labels_500Response(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	gh := NewGithub(ts.URL, git.NewMock(), "", cache.NewMockAidyCache())
	gh.ch.WithRemote("valid-repo")

	labels, err := gh.Labels()

	require.Error(t, err, "Labels should return an error for non-200 response")
	assert.Contains(t, err.Error(), "cannot retrieve labels using the following url")
	assert.Contains(t, err.Error(), "response: '500 Internal Server Error'")
	assert.Nil(t, labels, "Labels should be nil on error")
}
func TestRealGithub_Labels_Non404Response(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	gh := NewGithub(ts.URL, git.NewMock(), "", cache.NewMockAidyCache())
	gh.ch.WithRemote("valid-repo")

	labels, err := gh.Labels()

	require.NoError(t, err, "Labels should not return an error for 404 response")
	assert.Empty(t, labels, "Labels should be empty on 404 response")
}

func TestRealGithub_Labels_UnmarshalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("invalid json")); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()
	gh := NewGithub(ts.URL, git.NewMock(), "", cache.NewMockAidyCache())
	gh.ch.WithRemote("valid-repo")

	labels, err := gh.Labels()

	require.Error(t, err, "Labels should return an error for JSON unmarshal failure")
	assert.Contains(t, err.Error(), "error unmarshaling issue json")
	assert.Nil(t, labels, "Labels should be nil on error")
}

func TestRealGithub_Remotes_Error(t *testing.T) {
	git := git.NewMockWithError(fmt.Errorf("simulated git error"))
	gh := NewGithub("", git, "", cache.NewMockAidyCache())

	remotes, err := gh.Remotes()

	require.Error(t, err, "Remotes should return an error when git remotes retrieval fails")
	assert.Contains(t, err.Error(), "cannot retrieve git remotes")
	assert.Contains(t, err.Error(), "simulated git error")
	assert.Nil(t, remotes, "Remotes should be nil on error")
}
