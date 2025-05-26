package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const example = `
model: 4o
openai-api-key: secret-key
`

func TestAider_OpenAiKey(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		require.NoError(t, os.RemoveAll(tmp), "Failed to remove temp dir")
	}()
	path := tmp + "/config.yml"
	content := []byte(example)
	err = os.WriteFile(path, content, 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	apiKey, err := config.OpenAiKey()

	require.NoError(t, err, "Error should be nil")
	assert.Equal(t, "secret-key", apiKey, "API key should match")
}

func TestAider_Model(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		require.NoError(t, os.RemoveAll(tmp), "Failed to remove temp dir")
	}()
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	apiKey, err := config.Model()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "4o", apiKey, "Model should match")
}

func TestAider_DeepseekKey(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		require.NoError(t, os.RemoveAll(tmp), "Failed to remove temp dir")
	}()
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	apiKey, err := config.DeepseekKey()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "unknown", apiKey, "Deepseek key should match")
}

func TestAider_GithubKey(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		require.NoError(t, os.RemoveAll(tmp), "Failed to remove temp dir")
	}()
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	key, err := config.GithubKey()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "", key, "Github key should match")
}

func TestAider_UnexistingFile(t *testing.T) {
	_, err := NewAider("nonexistent.yml")
	assert.Error(t, err, "Should return an error for nonexistent file")
}

func TestAider_InvalidYaml(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		require.NoError(t, os.RemoveAll(tmp), "Failed to remove temp dir")
	}()
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(`"strange/format'''`), 0644)
	require.NoError(t, err, "Failed to write config file")

	_, err = NewAider(path)

	assert.Error(t, err, "Failed to load config")
}
