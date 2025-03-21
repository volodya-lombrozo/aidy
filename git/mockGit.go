package git

type MockGit struct{}

func (m *MockGit) GetBaseBranchName() (string, error) {
    // For mock purposes, assume 'main' is the base branch
    return "main", nil
}

func (m *MockGit) GetBranchName() (string, error) {
    return "main", nil
}
func (m *MockGit) GetDiff(baseBranch string) (string, error) {
    return "mock-diff", nil
}
