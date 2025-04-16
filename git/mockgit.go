package git

type MockGit struct{}

func (m *MockGit) GetBaseBranchName() (string, error) {
	return "main", nil
}

func (m *MockGit) AppendToCommit() error {
	return nil
}

func (m *MockGit) GetBranchName() (string, error) {
	return "41_working_branch", nil
}

func (m *MockGit) GetDiff() (string, error) {
	return "mock-diff", nil
}

func (m *MockGit) GetCurrentDiff() (string, error) {
	return "current-mock-diff", nil
}

func (m *MockGit) CommitChanges(messages ...string) error {
	return nil
}

func (m *MockGit) GetCurrentCommitMessage() (string, error) {
	return "feat(#42): current commit message", nil
}

func (r *MockGit) Remotes() ([]string, error) {
	return []string{"https://github.com/volodya-lombrozo/aidy.git", "https://github.com/volodya-lombrozo/forked-aidy.git"}, nil
}
