package git

import (
	"fmt"
	"log"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/executor"
)

type mock struct {
	shell executor.Executor
	dir   string
	err   error
}

func NewMock() Git {
	return &mock{dir: "/dev/null", shell: executor.NewMock()}
}

func NewMockWithDir(dir string) Git {
	return &mock{dir: dir, shell: executor.NewMock(), err: nil}
}

func NewMockWithShell(shell executor.Executor) Git {
	return &mock{dir: "/dev/null", shell: shell, err: nil}
}

func NewMockWithError(err error) Git {
	return &mock{dir: "/dev/null", shell: executor.NewMock(), err: err}
}

func (m *mock) Run(args ...string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	out, err := m.shell.RunCommand("git", args...)
	if err != nil {
		return out, err
	}
	return out, nil
}

func (m *mock) Log(since string) ([]string, error) {
	_, err := m.Run("log", fmt.Sprintf("%s..HEAD", since), "--pretty=format:%s")
	if err != nil {
		return nil, err
	}
	return []string{
		"ci(#120): Update CI to use Ubuntu 24.04 and add .aidy to gitignore",
		"chore(deps): update dependency ruby to v3.4.3 (#117)",
	}, nil
}

func (m *mock) Tags(repo string) ([]string, error) {
	res, err := m.Run("fetch", repo, "--tags")
	if err != nil {
		log.Printf("Error fetching tags from repository %s: %v", repo, err)
		return nil, err
	}
	if strings.Contains(res, "absent") {
		log.Printf("No tags found in repository %s", repo)
		return []string{}, nil
	}

	log.Printf("Fetched tags from repository %s: %s", repo, res)
	return []string{"v1.0", "v2.0"}, nil
}

func (m *mock) Amend(message string) error {
	if _, err := m.shell.RunCommand("git commit --amend -m " + message); err != nil {
		return err
	}
	return nil
}

func (m *mock) AddAll() error {
	if _, err := m.shell.RunCommand("git add --all"); err != nil {
		return err
	}
	return nil
}

func (m *mock) BaseBranch() (string, error) {
	return "main", nil
}

func (m *mock) Reset(ref string) error {
	if _, err := m.shell.RunCommand("git reset --soft", ref); err != nil {
		return err
	}
	return nil
}

func (m *mock) Append() error {
	if _, err := m.shell.RunCommand("git commit --amend --no-edit"); err != nil {
		return err
	}
	return nil
}

func (m *mock) CurrentBranch() (string, error) {
	return "41_working_branch", nil
}

func (m *mock) Diff() (string, error) {
	return "mock-diff", m.err
}

func (m *mock) CurrentDiff() (string, error) {
	return "current-mock-diff", nil
}

func (m *mock) CommitMessage() (string, error) {
	return "feat(#42): current commit message", nil
}

func (r *mock) Remotes() ([]string, error) {
	return []string{"https://github.com/volodya-lombrozo/aidy.git", "https://github.com/volodya-lombrozo/forked-aidy.git"}, nil
}

func (r *mock) Installed() (bool, error) {
	return true, nil
}

func (r *mock) Root() (string, error) {
	return r.dir, r.err
}

func (r *mock) Checkout(branch string) error {
	if _, err := r.shell.RunCommand("git checkout -b " + branch); err != nil {
		return err
	}
	return nil
}
