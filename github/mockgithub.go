package github

type MockGithub struct{}

func (m *MockGithub) IssueDescription(number string) string {
	return "Mock description for issue #" + number
}

func (m *MockGithub) Labels() []string {
	return []string{"bug", "documentation", "question"}
}
