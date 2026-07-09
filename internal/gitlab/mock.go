package gitlab

import "fmt"

type MockGitlab struct {
	Error error
}

func NewMock() *MockGitlab {
	return &MockGitlab{Error: nil}
}

func (m *MockGitlab) MergeRequestByBranch(branch string) (string, string, error) {
	return fmt.Sprintf("mock title for branch '%s'", branch), fmt.Sprintf("mock body for branch '%s'", branch), m.Error
}
