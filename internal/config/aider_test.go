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
	tmp := t.TempDir()
	path := tmp + "/config.yml"
	content := []byte(example)
	err := os.WriteFile(path, content, 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	apiKey, err := config.OpenAiKey()

	require.NoError(t, err, "Error should be nil")
	assert.Equal(t, "secret-key", apiKey, "API key should match")
}

func TestAider_Model(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.yml"
	err := os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	apiKey, err := config.Model()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "4o", apiKey, "Model should match")
}

func TestAider_DeepseekKey(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.yml"
	err := os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	apiKey, err := config.DeepseekKey()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "unknown", apiKey, "Deepseek key should match")
}

func TestAider_GithubKey(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.yml"
	err := os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	key, err := config.GithubKey()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "", key, "Github key should match")
}

func TestAider_Provider(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.yml"
	err := os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	provider, err := config.Provider()

	require.NoError(t, err, "Error should be nil")
	assert.Equal(t, "openai", provider, "Provider should match")
}

func TestAider_Token(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.yml"
	err := os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := NewAider(path)
	require.NoError(t, err, "Failed to load config")

	token, err := config.Token()

	require.NoError(t, err, "Error should be nil")
	assert.Equal(t, "secret-key", token, "Token should match")
}

func TestAider_UnexistingFile(t *testing.T) {
	_, err := NewAider("nonexistent.yml")
	assert.Error(t, err, "Should return an error for nonexistent file")
}

func TestAider_InvalidYaml(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/config.yml"
	err := os.WriteFile(path, []byte(`"strange/format'''`), 0644)
	require.NoError(t, err, "Failed to write config file")

	_, err = NewAider(path)

	assert.Error(t, err, "Failed to load config")
}
