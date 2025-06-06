package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/git"
)

const aidyconf = `
default-model: test-model
api-keys:
  openai: test-openai-key
  github: test-github-key
  deepseek: test-deepseek-key
models:
  test-model:
    provider: deepseek
    model-id: test-model-1
`

const aider = `
model: gpt-4o-test
openai-api-key: test-openai-key
`

func TestCascade_AidyFirst(t *testing.T) {
	original, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")
	tmp := t.TempDir()
	tmpConfFile(t, tmp, ".aidy.conf.yaml", aidyconf)
	defer func() {
		err := os.Chdir(original)
		require.NoError(t, err, "Failed to change back to original directory")
		err = os.RemoveAll(tmp)
		require.NoError(t, err, "Failed to remove temp directory")
	}()
	err = os.Chdir(tmp)
	require.NoError(t, err, "Failed to change directory to temp dir")
	conf, err := NewCascade(git.NewMock())

	require.NoError(t, err, "Failed to create cascade config")
	openAIKey, err := conf.OpenAiKey()
	assert.NoError(t, err)
	assert.Equal(t, "test-openai-key", openAIKey)

	githubKey, err := conf.GithubKey()
	assert.NoError(t, err)
	assert.Equal(t, "test-github-key", githubKey)

	deepseekKey, err := conf.DeepseekKey()
	assert.NoError(t, err)
	assert.Equal(t, "test-deepseek-key", deepseekKey)

	model, err := conf.Model()
	assert.NoError(t, err)
	assert.Equal(t, "test-model-1", model)

	token, err := conf.Token()
	assert.NoError(t, err)
	assert.Equal(t, "test-deepseek-key", token, "Token should match")

	provider, err := conf.Provider()
	assert.NoError(t, err)
	assert.Equal(t, "deepseek", provider, "Provider should match")
}

func TestCascade_AiderFirst(t *testing.T) {
	original, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")
	tmp := t.TempDir()
	tmpConfFile(t, tmp, ".aider.conf.yaml", aider)
	defer func() {
		err := os.Chdir(original)
		require.NoError(t, err, "Failed to change back to original directory")
		err = os.RemoveAll(tmp)
		require.NoError(t, err, "Failed to remove temp directory")
	}()
	err = os.Chdir(tmp)
	require.NoError(t, err, "Failed to change directory to temp dir")
	folder := func() (string, error) { return tmp, nil }
	conf, err := NewCascadeInDirs(folder)
	require.NoError(t, err, "Failed to create cascade config")

	openai, err := conf.OpenAiKey()
	assert.NoError(t, err)
	assert.Equal(t, "test-openai-key", openai)

	github, err := conf.GithubKey()
	assert.NoError(t, err)
	assert.Equal(t, "", github)

	deepseek, err := conf.DeepseekKey()
	assert.NoError(t, err)
	assert.Equal(t, "unknown", deepseek)

	model, err := conf.Model()
	assert.NoError(t, err)
	assert.Equal(t, "gpt-4o-test", model)

	token, err := conf.Token()
	assert.NoError(t, err)
	assert.Equal(t, "test-openai-key", token, "Token should match")

	token, err = conf.Provider()
	assert.NoError(t, err)
	assert.Equal(t, "openai", token, "Provider should match")
}

func tmpConfFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "Failed to write temp config file")
	return path
}
