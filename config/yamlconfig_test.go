package config

import (
    "io/ioutil"
    "os"
    "testing"
)

func TestYAMLConfig_GetOpenAIAPIKey(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "configtest")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)

    configFilePath := tempDir + "/config.yml"
    configContent := []byte("openai-api-key: test-api-key")
    err = ioutil.WriteFile(configFilePath, configContent, 0644)
    if err != nil {
        t.Fatalf("Failed to write config file: %v", err)
    }

    yamlConfig, err := NewYAMLConfig(configFilePath)
    if err != nil {
        t.Fatalf("Failed to create YAMLConfig: %v", err)
    }

    apiKey, err := yamlConfig.GetOpenAIAPIKey()
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if apiKey != "test-api-key" {
        t.Fatalf("Expected API key 'test-api-key', got '%s'", apiKey)
    }
}

func TestYAMLConfig_GetOpenAIAPIKey_MissingKey(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "configtest")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)

    configFilePath := tempDir + "/config.yml"
    configContent := []byte("openai-api-key: ")
    err = ioutil.WriteFile(configFilePath, configContent, 0644)
    if err != nil {
        t.Fatalf("Failed to write config file: %v", err)
    }

    yamlConfig, err := NewYAMLConfig(configFilePath)
    if err != nil {
        t.Fatalf("Failed to create YAMLConfig: %v", err)
    }

    _, err = yamlConfig.GetOpenAIAPIKey()
    if err == nil {
        t.Fatal("Expected error for missing API key, got none")
    }
}
