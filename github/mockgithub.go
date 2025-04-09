package github

type MockGithub struct{}

func (m *MockGithub) IssueDescription(number string) string {
	return "Mock description for issue #" + number
}
