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

func (r *MockGit) Installed() (bool, error) {
	return true, nil
}

func (r *MockGit) Root() (string, error) {
	return "/dev/null", nil
}

type MockGitWithDir struct {
	dir string
}

func NewMockGitWithDir(dir string) *MockGitWithDir {
	return &MockGitWithDir{dir: dir}
}

func (m *MockGitWithDir) GetBaseBranchName() (string, error) {
	return "main", nil
}

func (m *MockGitWithDir) AppendToCommit() error {
	return nil
}

func (m *MockGitWithDir) GetBranchName() (string, error) {
	return "41_working_branch", nil
}

func (m *MockGitWithDir) GetDiff() (string, error) {
	return "mock-diff", nil
}

func (m *MockGitWithDir) GetCurrentDiff() (string, error) {
	return "current-mock-diff", nil
}

func (m *MockGitWithDir) CommitChanges(messages ...string) error {
	return nil
}

func (m *MockGitWithDir) GetCurrentCommitMessage() (string, error) {
	return "feat(#42): current commit message", nil
}

func (m *MockGitWithDir) Remotes() ([]string, error) {
	return []string{"https://github.com/volodya-lombrozo/aidy.git", "https://github.com/volodya-lombrozo/forked-aidy.git"}, nil
}

func (m *MockGitWithDir) Installed() (bool, error) {
	return true, nil
}

func (m *MockGitWithDir) Root() (string, error) {
	return m.dir, nil
}
