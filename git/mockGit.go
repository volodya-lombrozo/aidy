package git

type MockGit struct{}

func (m *MockGit) GetBranchName() (string, error) {
    return "mock-branch-name", nil
}
