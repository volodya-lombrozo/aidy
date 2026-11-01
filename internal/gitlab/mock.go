package gitlab

import "fmt"

type MockGitlab struct {
	Error error
}

func NewMock() *MockGitlab {
	return &MockGitlab{Error: nil}
}

func (m *MockGitlab) Description(number string, repository string) (string, error) {
	return fmt.Sprintf("mock description for issue '#%s' in '%s'", number, repository), m.Error
}

func (m *MockGitlab) MergeRequestByBranch(branch string) (string, string, error) {
	return fmt.Sprintf("mock title for branch '%s'", branch), fmt.Sprintf("mock body for branch '%s'", branch), m.Error
}
