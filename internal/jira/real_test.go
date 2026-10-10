package jira

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealJira_RetrievesIssueType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/issue/PROJ-123", r.URL.Path)
		assert.Equal(t, "issuetype", r.URL.Query().Get("fields"))
		if _, err := w.Write([]byte(`{"key":"PROJ-123","fields":{"issuetype":{"name":"Bug"}}}`)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()

	kind, err := NewJira(ts.URL+"/", "", "token").IssueType("PROJ-123")

	require.NoError(t, err)
	assert.Equal(t, "Bug", kind)
}

func TestRealJira_UsesBasicAuthWithEmail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok, "basic auth should be used when email is set")
		assert.Equal(t, "me@example.com", user)
		assert.Equal(t, "token", pass)
		if _, err := w.Write([]byte(`{"fields":{"issuetype":{"name":"Task"}}}`)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()

	_, err := NewJira(ts.URL, "me@example.com", "token").IssueType("PROJ-1")

	require.NoError(t, err)
}

func TestRealJira_UsesBearerTokenWithoutEmail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		if _, err := w.Write([]byte(`{"fields":{"issuetype":{"name":"Task"}}}`)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()

	_, err := NewJira(ts.URL, "", "token").IssueType("PROJ-1")

	require.NoError(t, err)
}

func TestRealJira_FailsOnUnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	_, err := NewJira(ts.URL, "", "token").IssueType("PROJ-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestRealJira_FailsOnInvalidJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`not json`)); err != nil {
			t.Errorf("Error writing response: %v", err)
		}
	}))
	defer ts.Close()

	_, err := NewJira(ts.URL, "", "token").IssueType("PROJ-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshaling jira issue json")
}

func TestOfflineJira_AlwaysFails(t *testing.T) {
	_, err := NewOffline().IssueType("PROJ-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "jira connection is not configured")
}

func TestMockJira_ReturnsType(t *testing.T) {
	kind, err := NewMock("Test").IssueType("PROJ-1")

	require.NoError(t, err)
	assert.Equal(t, "Test", kind)
}
