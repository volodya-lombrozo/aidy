package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMock_GithubKey(t *testing.T) {
	expected := "github-api-key"
	conf := NewMock()
	conf.MockGithub = expected

	key, err := conf.GithubKey()

	require.NoError(t, err, "Expected no error when getting GitHub API key")
	assert.Equal(t, expected, key, "Expected GitHub API key to match the mock value")
}

func TestMock_Model(t *testing.T) {
	expected := "gpt-4o"
	conf := NewMock()
	conf.MockModel = expected

	model, err := conf.Model()

	require.NoError(t, err, "Expected no error when getting model")
	assert.Equal(t, expected, model, "Expected model to match the mock value")
}

func TestMock_Provider(t *testing.T) {
	expected := "openai"
	conf := NewMock()
	conf.MockProvider = expected

	provider, err := conf.Provider()

	require.NoError(t, err, "Expected no error when getting provider")
	assert.Equal(t, expected, provider, "Expected provider to match the mock value")
}

func TestMock_Token(t *testing.T) {
	expected := "mock-token"
	conf := NewMock()
	conf.MockToken = expected

	token, err := conf.Token()

	require.NoError(t, err, "Expected no error when getting token")
	assert.Equal(t, expected, token, "Expected token to match the mock value")
}
