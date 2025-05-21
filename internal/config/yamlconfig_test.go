package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const FULL = `
# Global settings
default-model: 4o

# API keys per provider
api-keys:
  openai: sk-...
  deepseek: ds-...

# Model definitions
models:
  4o:
    provider: openai
    model-id: gpt-4o
    max-tokens: 8192
    temperature: 0.7
    use-streaming: true

  4o-mini:
    provider: openai
    model-id: gpt-4o-mini
    max-tokens: 4096
    temperature: 0.5
    use-streaming: false

  deepseek:
    provider: deepseek
    model-id: deepseek-chat
    max-tokens: 6000
    temperature: 0.8
    use-streaming: true
    custom-option: experimental-mode`

const MODEL_ONLY = `
default-model: 4o
models:
  4o:
    model-id: gpt-4o
`

const KEYS = `
api-keys:
  openai: sk-test
  github: gh-test
`

func TestNewConfig(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err)
	defer clean(t, tmp)
	configFilePath := tmp + "/config.yml"
	configContent := []byte(FULL)
	err = os.WriteFile(configFilePath, configContent, 0644)
	require.NoError(t, err)

	config := YamlConf(configFilePath)

	assert.Equal(t, "4o", config.DefaultModel)
	assert.Equal(t, "sk-...", config.APIKeys["openai"])
	assert.Equal(t, "ds-...", config.APIKeys["deepseek"])
	model4o := config.Models["4o"]
	assert.Equal(t, "openai", model4o["provider"])
	assert.Equal(t, "gpt-4o", model4o["model-id"])
	assert.Equal(t, "8192", model4o["max-tokens"])
	assert.Equal(t, "0.7", model4o["temperature"])
	assert.Equal(t, "true", model4o["use-streaming"])
	model4oMini := config.Models["4o-mini"]
	assert.Equal(t, "openai", model4oMini["provider"])
	assert.Equal(t, "gpt-4o-mini", model4oMini["model-id"])
	assert.Equal(t, "4096", model4oMini["max-tokens"])
	assert.Equal(t, "0.5", model4oMini["temperature"])
	assert.Equal(t, "false", model4oMini["use-streaming"])
	modelDeepseek := config.Models["deepseek"]
	assert.Equal(t, "deepseek", modelDeepseek["provider"])
	assert.Equal(t, "deepseek-chat", modelDeepseek["model-id"])
	assert.Equal(t, "6000", modelDeepseek["max-tokens"])
	assert.Equal(t, "0.8", modelDeepseek["temperature"])
	assert.Equal(t, "true", modelDeepseek["use-streaming"])
	assert.Equal(t, "experimental-mode", modelDeepseek["custom-option"])
}

func TestGetModel(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err)
	defer clean(t, tmp)
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(MODEL_ONLY), 0644)
	require.NoError(t, err)

	config := YamlConf(path)

	model, err := config.GetModel()
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "gpt-4o", model, "Model ID should match")
}

func TestGetOpenAIAPIKey(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temp dir")
	defer clean(t, tmp)
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(KEYS), 0644)
	require.NoError(t, err, "Failed to write config file")

	config := YamlConf(path)

	apiKey, err := config.GetOpenAIAPIKey()
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "sk-test", apiKey, "API key should match")
}

func TestGetGithubAPIKey(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temo dir")
	defer clean(t, tmp)
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(KEYS), 0644)
	require.NoError(t, err, "Filed to write config file")

	config := YamlConf(path)

	apiKey, err := config.GetGithubAPIKey()
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "gh-test", apiKey, "API key should match")
}

func clean(t *testing.T, tmp string) {
	if err := os.RemoveAll(tmp); err != nil {
		t.Fatalf("Error removing temp directory: %v", err)
	}
}
