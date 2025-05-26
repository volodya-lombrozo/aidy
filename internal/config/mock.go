package config

type MockConfig struct {
	MockDeepseek string
	MockOpenai   string
	MockGithub   string
	MockModel    string
}

func NewMock() *MockConfig {
	return &MockConfig{
		MockOpenai:   "mock-openai-key",
		MockGithub:   "mock-github-key",
		MockDeepseek: "mock-deepseek-key",
		MockModel:    "gpt-4o",
	}
}

func (m *MockConfig) DeepseekKey() (string, error) {
	return m.MockDeepseek, nil
}

func (m *MockConfig) GithubKey() (string, error) {
	return m.MockGithub, nil
}

func (m *MockConfig) Model() (string, error) {
	return m.MockModel, nil
}

func (m *MockConfig) OpenAiKey() (string, error) {
	return m.MockOpenai, nil
}
