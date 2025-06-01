package github

type MockGithub struct{}

func NewMock() *MockGithub {
	return &MockGithub{}
}

func (m *MockGithub) Description(number string) (string, error) {
	return "Mock description for issue #" + number, nil
}

func (m *MockGithub) Labels() ([]string, error) {
	return []string{"bug", "documentation", "question"}, nil
}

func (m *MockGithub) Remotes() ([]string, error) {
	return []string{"volodya-lombrozo/aidy", "volodya-lombrozo/jtcop"}, nil
}
