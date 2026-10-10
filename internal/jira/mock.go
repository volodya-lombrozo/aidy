package jira

type MockJira struct {
	Type  string
	Error error
}

func NewMock(kind string) *MockJira {
	return &MockJira{Type: kind}
}

func (m *MockJira) IssueType(key string) (string, error) {
	return m.Type, m.Error
}
