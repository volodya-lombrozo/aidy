package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const example = `
model: 4o
openai-api-key: secret-key
`

func TestAiderGetOpenAIAPIKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "configtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Error removing temp directory: %v", err)
		}
	}()

	configFilePath := tempDir + "/config.yml"
	configContent := []byte(example)
	err = os.WriteFile(configFilePath, configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config := NewAiderConf(configFilePath)

	apiKey, err := config.GetOpenAIAPIKey()
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "secret-key", apiKey, "API key should match")
}

func TestAiderGetModel(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "configtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Error removing temp directory: %v", err)
		}
	}()

	configFilePath := tempDir + "/config.yml"
	configContent := []byte(example)
	err = os.WriteFile(configFilePath, configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config := NewAiderConf(configFilePath)

	apiKey, err := config.GetModel()
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, "4o", apiKey, "Model should match")
}
