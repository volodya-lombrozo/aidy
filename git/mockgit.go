package git

type MockGit struct{}

func (m *MockGit) GetBaseBranchName() (string, error) {
	return "main", nil
}

func (m *MockGit) AppendToCommit() error {
	// Mock implementation, no actual git operations
	return nil
}

func (m *MockGit) GetBranchName() (string, error) {
	return "41_working_branch", nil
}
func (m *MockGit) GetDiff() (string, error) {
	return "mock-diff", nil
}

func (m *MockGit) CommitChanges() error {
	// Mock implementation, no actual git operations
	return nil
}

func (m *MockGit) GetCurrentCommitMessage() (string, error) {
	return "feat(#42): current commit message", nil
}
