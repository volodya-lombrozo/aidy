package config

type MockConfig struct {
	MockDeepseek string
	MockOpenai   string
	MockGithub   string
	MockModel    string
	Error        error
	MockToken    string
	MockProvider string
}

func NewMock() *MockConfig {
	return &MockConfig{
		MockOpenai:   "mock-openai-key",
		MockGithub:   "mock-github-key",
		MockDeepseek: "mock-deepseek-key",
		MockModel:    "gpt-4o",
		MockToken:    "mock-token",
		MockProvider: "openai",
	}
}

func (m *MockConfig) DeepseekKey() (string, error) {
	return m.MockDeepseek, m.Error
}

func (m *MockConfig) GithubKey() (string, error) {
	return m.MockGithub, m.Error
}

func (m *MockConfig) Model() (string, error) {
	return m.MockModel, m.Error
}

func (m *MockConfig) OpenAiKey() (string, error) {
	return m.MockOpenai, m.Error
}

func (m *MockConfig) Provider() (string, error) {
	return m.MockProvider, m.Error
}

func (m *MockConfig) Token() (string, error) {
	return m.MockToken, m.Error
}
