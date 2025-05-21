package config

type MockConfig struct {
	OpenAIAPIKey string
	GitHubAPIKey string
}

func (m *MockConfig) GetModel() (string, error) {
	return "gpt-4o", nil
}

func NewMockConfig(apiKey string, github string) *MockConfig {
	return &MockConfig{OpenAIAPIKey: apiKey, GitHubAPIKey: github}
}

func (m *MockConfig) GetOpenAIAPIKey() (string, error) {
	return m.OpenAIAPIKey, nil
}

func (m *MockConfig) GetGithubAPIKey() (string, error) {
	return m.GitHubAPIKey, nil
}
