package config

type MockConfig struct {
	OpenAIAPIKey string
}

func NewMockConfig(apiKey string) *MockConfig {
	return &MockConfig{OpenAIAPIKey: apiKey}
}

func (m *MockConfig) GetOpenAIAPIKey() (string, error) {
	return m.OpenAIAPIKey, nil
}
