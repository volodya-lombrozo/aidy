package config

import (
	"testing"
)

func TestMockConfig_GetOpenAIAPIKey(t *testing.T) {
	expectedAPIKey := "test-api-key"
	mockConfig := NewMockConfig(expectedAPIKey, "")

	apiKey, err := mockConfig.GetOpenAIAPIKey()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if apiKey != expectedAPIKey {
		t.Fatalf("Expected API key '%s', got '%s'", expectedAPIKey, apiKey)
	}
}

func TestMockConfig_GetGithubAPIKey(t *testing.T) {
	expectedAPIKey := "github-api-key"
	mockConfig := NewMockConfig("", expectedAPIKey)
	apiKey, err := mockConfig.GetGithubAPIKey()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if apiKey != expectedAPIKey {
		t.Fatalf("Expected API key '%s', got '%s'", expectedAPIKey, apiKey)
	}
}
