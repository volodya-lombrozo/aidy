package github

type MockGithub struct{}

func (m *MockGithub) Description(number string) string {
	return "Mock description for issue #" + number
}

func (m *MockGithub) Labels() []string {
	return []string{"bug", "documentation", "question"}
}

func (m *MockGithub) Remotes() []string {
	return []string{"volodya-lombrozo/aidy", "volodya-lombrozo/jtcop"}
}
