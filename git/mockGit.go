package git

type MockGit struct{}

func (m *MockGit) GetBranchName() (string, error) {
    return "mock-branch-name", nil
func (m *MockGit) GetDiff(baseBranch string) (string, error) {
    return "mock-diff", nil
}
