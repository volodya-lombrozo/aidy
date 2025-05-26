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

func TestAiderGetOpenAIAPIKey(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		require.NoError(t, os.RemoveAll(tmp), "Failed to remove temp dir")
	}()
	path := tmp + "/config.yml"
	content := []byte(example)
	err = os.WriteFile(path, content, 0644)
	require.NoError(t, err, "Failed to write config file")

	config := NewAider(path)

	apiKey, err := config.OpenAiKey()
	require.NoError(t, err, "Error should be nil")
	assert.Equal(t, "secret-key", apiKey, "API key should match")
}

func TestAiderGetModel(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		require.NoError(t, os.RemoveAll(tmp), "Failed to remove temp dir")
	}()
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(example), 0644)
	require.NoError(t, err, "Failed to write config file")

	config := NewAider(path)

	apiKey, err := config.Model()
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "4o", apiKey, "Model should match")
}
