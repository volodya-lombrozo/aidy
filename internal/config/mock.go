package config

type MockConfig struct {
	MockGithub   string
	MockModel    string
	Error        error
	MockToken    string
	MockProvider string
}

func NewMock() *MockConfig {
	return &MockConfig{
		MockGithub:   "mock-github-key",
		MockModel:    "gpt-4o",
		MockToken:    "mock-token",
		MockProvider: "openai",
	}
}

func (m *MockConfig) GithubKey() (string, error) {
	return m.MockGithub, m.Error
}

func (m *MockConfig) Model() (string, error) {
	return m.MockModel, m.Error
}

func (m *MockConfig) Provider() (string, error) {
	return m.MockProvider, m.Error
}

func (m *MockConfig) Token() (string, error) {
	return m.MockToken, m.Error
}
