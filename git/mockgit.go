package git

import (
	"fmt"

	"github.com/volodya-lombrozo/aidy/executor"
)

type MockGit struct {
	Shell executor.Executor
}

func (m *MockGit) Log(since string) ([]string, error) {
	if m.Shell != nil {
		if _, err := m.Shell.RunCommand(fmt.Sprintf("git log %s..HEAD --pretty=format:%%s", since)); err != nil {
			return nil, err
		}
	}
	return []string{"ci(#120): Update CI to use Ubuntu 24.04 and add .aidy to gitignore", "chore(deps): update dependency ruby to v3.4.3 (#117)"}, nil
}

func (m *MockGit) AddTag(tag string, message string) error {
	if m.Shell != nil {
		if _, err := m.Shell.RunCommand(fmt.Sprintf("git tag -a %s -m %s", tag, message)); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockGit) AddTagCommand(tag string, message string) string {
	return fmt.Sprintf("git tag -a \"%s\" -m \"%s\"", tag, message)
}

func (m *MockGit) Tags() ([]string, error) {
	if m.Shell != nil {
		if _, err := m.Shell.RunCommand("git tag"); err != nil {
			return nil, err
		}
	}
	return []string{"v1.0", "v2.0"}, nil
}

func (m *MockGit) Amend(message string) error {
	if m.Shell != nil {
		if _, err := m.Shell.RunCommand("git commit --amend -m " + message); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockGit) AddAll() error {
	if m.Shell != nil {
		if _, err := m.Shell.RunCommand("git add --all"); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockGit) GetBaseBranchName() (string, error) {
	return "main", nil
}

func (m *MockGit) Reset(ref string) error {
	if m.Shell != nil {
		if _, err := m.Shell.RunCommand("git reset --soft", ref); err != nil {
			return err
		}
	}
	return nil
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

func (r *MockGit) Checkout(branch string) error {
	if r.Shell != nil {
		if _, err := r.Shell.RunCommand("git checkout -b " + branch); err != nil {
			return err
		}
	}
	return nil
}

type MockGitWithDir struct {
	dir string
}

func NewMockGitWithDir(dir string) Git {
	return &MockGitWithDir{dir: dir}
}

func (m *MockGitWithDir) AddTagCommand(tag string, message string) string {
	return fmt.Sprintf("git tag -a \"%s\" -m \"%s\"", tag, message)
}

func (m *MockGitWithDir) Log(since string) ([]string, error) {
	return []string{"ci(#120): Update CI to use Ubuntu 24.04 and add .aidy to gitignore", "chore(deps): update dependency ruby to v3.4.3 (#117)"}, nil
}

func (m *MockGitWithDir) AddTag(tag string, message string) error {
	return nil
}

func (m *MockGitWithDir) Tags() ([]string, error) {
	return []string{"v1.0", "v2.0"}, nil
}
func (m *MockGitWithDir) Reset(ref string) error {
	return nil
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

func (m *MockGitWithDir) AddAll() error {
	return nil
}

func (m *MockGitWithDir) Amend(message string) error {
	panic("unimplemented")
}

// Checkout implements Git.
func (m *MockGitWithDir) Checkout(branch string) error {
	panic("unimplemented")
}
