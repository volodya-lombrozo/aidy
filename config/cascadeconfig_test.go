package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volodya-lombrozo/aidy/git"
)

func createTempConfigFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	return path
}

func TestCascadeConfig(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	tempDir := t.TempDir()
	aidyConfContent := `
default-model: test-model
api-keys:
  openai: test-openai-key
  github: test-github-key
  deepseek: test-deepseek-key
models:
  test-model:
    provider: mock
    model-id: test-model-1
`
	createTempConfigFile(t, tempDir, ".aidy.conf.yaml", aidyConfContent)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change back to original directory: %v", err)
		}
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Error removing temp directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	mockGit := git.NewMock()

	conf := NewCascadeConfig(mockGit)

	openAIKey, err := conf.GetOpenAIAPIKey()
	assert.NoError(t, err)
	assert.Equal(t, "test-openai-key", openAIKey)

	githubKey, err := conf.GetGithubAPIKey()
	assert.NoError(t, err)
	assert.Equal(t, "test-github-key", githubKey)

	deepseekKey, err := conf.GetDeepseekAPIKey()
	assert.NoError(t, err)
	assert.Equal(t, "test-deepseek-key", deepseekKey)

	model, err := conf.GetModel()
	assert.NoError(t, err)
	assert.Equal(t, "test-model-1", model)
}
