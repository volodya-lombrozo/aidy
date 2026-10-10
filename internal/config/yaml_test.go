package config

import (
	"os"
	"path/filepath"
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

	config, err := YamlConf(configFilePath)

	require.NoError(t, err, "no error expected")
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

func TestYamlConfFileReadError(t *testing.T) {
	_, err := YamlConf("/non/existent/path/config.yml")
	assert.Error(t, err, "Expected an error when the file cannot be read")
}

func TestYamlConfUnmarshalError(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err)
	defer clean(t, tmp)
	path := tmp + "/invalid_config.yml"
	content := []byte("invalid_yaml: : :")
	err = os.WriteFile(path, content, 0644)
	require.NoError(t, err, "Failed to write invalid config file")

	_, err = YamlConf(path)

	assert.Error(t, err, "Expected an error when the file cannot be unmarshaled")
}

func TestYamlConfModelOnly(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err)
	defer clean(t, tmp)
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(MODEL_ONLY), 0644)
	require.NoError(t, err)
	config, err := YamlConf(path)
	require.NoError(t, err, "no error expected")

	model, err := config.Model()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "gpt-4o", model, "Model ID should match")
}

func TestGetGithubAPIKey(t *testing.T) {
	tmp, err := os.MkdirTemp("", "configtest")
	require.NoError(t, err, "Failed to create temo dir")
	defer clean(t, tmp)
	path := tmp + "/config.yml"
	err = os.WriteFile(path, []byte(KEYS), 0644)
	require.NoError(t, err, "Filed to write config file")
	config, err := YamlConf(path)
	require.NoError(t, err, "Failed to load config")

	key, err := config.GithubKey()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "gh-test", key, "API key should match")
}

func TestYaml_Provider(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yml")
	err := os.WriteFile(path, []byte(FULL), 0644)
	require.NoError(t, err, "Failed to write config file")

	config, err := YamlConf(path)

	require.NoError(t, err, "Failed to load config")
	provider, err := config.Provider()
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "openai", provider, "Provider should match")
}

func TestYaml_Model(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yml")
	err := os.WriteFile(path, []byte(FULL), 0644)
	require.NoError(t, err, "Failed to write config file")

	config, err := YamlConf(path)

	require.NoError(t, err, "Failed to load config")
	provider, err := config.Model()
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "gpt-4o", provider, "Provider should match")
}

func TestYaml_Token(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yml")
	err := os.WriteFile(path, []byte(FULL), 0644)
	require.NoError(t, err, "Failed to write config file")
	config, err := YamlConf(path)
	require.NoError(t, err, "Failed to load config")

	token, err := config.Token()

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "sk-...", token, "Token should match")
}

func TestWriteYaml(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yml")
	original := &YamlConfig{
		DefaultModel: "openai",
		APIKeys: map[string]string{
			"openai": "sk-test",
			"github": "gh-test",
		},
		Models: map[string]map[string]string{
			"openai": {
				"provider": "openai",
				"model-id": "gpt-4o",
			},
		},
	}

	err := WriteYaml(path, original)

	require.NoError(t, err, "no error expected")
	written, err := YamlConf(path)
	require.NoError(t, err, "written config should be readable")
	assert.Equal(t, "openai", written.DefaultModel)
	assert.Equal(t, "sk-test", written.APIKeys["openai"])
	assert.Equal(t, "gh-test", written.APIKeys["github"])
	assert.Equal(t, "gpt-4o", written.Models["openai"]["model-id"])
}

func TestWriteYaml_FileWriteError(t *testing.T) {
	err := WriteYaml("/non/existent/dir/config.yml", &YamlConfig{})

	assert.Error(t, err, "Expected an error when the file cannot be written")
}

func clean(t *testing.T, tmp string) {
	if err := os.RemoveAll(tmp); err != nil {
		t.Fatalf("Error removing temp directory: %v", err)
	}
}

func TestYamlConf_EnablesMetricsByDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(KEYS), 0644))
	config, err := YamlConf(path)
	require.NoError(t, err, "Failed to load config")

	enabled, err := config.Metrics()

	require.NoError(t, err)
	assert.True(t, enabled, "metrics should be enabled when not configured")
}

func TestYamlConf_DisablesMetrics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(KEYS+"metrics: false\n"), 0644))
	config, err := YamlConf(path)
	require.NoError(t, err, "Failed to load config")

	enabled, err := config.Metrics()

	require.NoError(t, err)
	assert.False(t, enabled, "metrics should be disabled by the configuration")
}

func TestWriteYaml_OmitsUnsetMetrics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")

	err := WriteYaml(path, &YamlConfig{DefaultModel: "4o"})

	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "metrics", "unset metrics option should not be written")
}

func TestYamlConf_ReadsJira(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	jira := "jira:\n  url: https://company.atlassian.net\n  email: me@example.com\n  token: jira-token\n"
	require.NoError(t, os.WriteFile(path, []byte(KEYS+jira), 0644))
	config, err := YamlConf(path)
	require.NoError(t, err, "Failed to load config")

	access, err := config.Jira()

	require.NoError(t, err)
	assert.Equal(t, Jira{URL: "https://company.atlassian.net", Email: "me@example.com", Token: "jira-token"}, access)
	assert.True(t, access.Configured(), "jira should be configured")
}

func TestYamlConf_LeavesJiraUnconfiguredByDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(KEYS), 0644))
	config, err := YamlConf(path)
	require.NoError(t, err, "Failed to load config")

	access, err := config.Jira()

	require.NoError(t, err)
	assert.False(t, access.Configured(), "jira should not be configured when missing")
}

func TestJira_RequiresUrlAndToken(t *testing.T) {
	assert.False(t, Jira{URL: "https://company.atlassian.net"}.Configured())
	assert.False(t, Jira{Token: "jira-token"}.Configured())
	assert.True(t, Jira{URL: "https://company.atlassian.net", Token: "jira-token"}.Configured())
}

func TestWriteYaml_OmitsUnsetJira(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")

	err := WriteYaml(path, &YamlConfig{DefaultModel: "4o"})

	require.NoError(t, err)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "jira", "unset jira option should not be written")
}
