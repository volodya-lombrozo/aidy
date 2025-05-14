package git

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/volodya-lombrozo/aidy/executor"
)

type RealGit struct {
	dir   string
	shell executor.Executor
}

func (r *RealGit) Amend(message string) error {
	_, err := r.shell.RunCommandInDir(r.dir, "git", "commit", "--amend", "-m", message)
	if err != nil {
		return fmt.Errorf("error amending commit: %w", err)
	}
	return nil
}

func NewRealGit(shell executor.Executor, dir ...string) *RealGit {
	var directory string
	if len(dir) > 0 && dir[0] != "" {
		directory = dir[0]
	} else {
		var err error
		directory, err = os.Getwd()
		if err != nil {
			panic(fmt.Errorf("failed to get current working directory: %w", err))
		}
	}
	return &RealGit{dir: directory, shell: shell}
}

func (r *RealGit) Reset(ref string) error {
	_, err := r.shell.RunCommandInDir(r.dir, "git", "reset", "--soft", ref)
	if err != nil {
		return fmt.Errorf("error resetting to %s: %w", ref, err)
	}
	return nil
}

func (r *RealGit) CommitChanges(messages ...string) error {
	_, err := r.shell.RunCommandInDir(r.dir, "git", "add", "--all")
	if err != nil {
		return fmt.Errorf("error adding changes: %w", err)
	}

	changed, err := r.shell.RunCommandInDir(r.dir, "git", "diff", "--name-only", "--cached")
	if err != nil {
		return fmt.Errorf("error getting changed files: %w", err)
	}

	var msg string
	if len(messages) > 0 {
		msg = messages[0]
	} else {
		msg = strings.TrimSpace("Committing changes to the following files:\n" + changed)
	}

	_, err = r.shell.RunCommandInDir(r.dir, "git", "commit", "-m", msg)
	if err != nil {
		return fmt.Errorf("error committing changes: %w", err)
	}

	return nil
}

func (r *RealGit) AppendToCommit() error {
	_, err := r.shell.RunCommandInDir(r.dir, "git", "add", "--all")
	if err != nil {
		return fmt.Errorf("error adding changes: %w", err)
	}
	_, err = r.shell.RunCommandInDir(r.dir, "git", "commit", "--amend", "--no-edit")
	if err != nil {
		return fmt.Errorf("error amending commit: %w", err)
	}
	return nil
}

func (r *RealGit) GetBranchName() (string, error) {
	branchName, err := r.shell.RunCommandInDir(r.dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	branchName = strings.TrimRight(branchName, "\r\n")
	return branchName, nil
}

func (r *RealGit) GetBaseBranchName() (string, error) {
	_, errMain := r.shell.RunCommandInDir(r.dir, "git", "show-ref", "--verify", "--quiet", "refs/heads/main")
	_, errMaster := r.shell.RunCommandInDir(r.dir, "git", "show-ref", "--verify", "--quiet", "refs/heads/master")
	if errMain == nil && errMaster == nil {
		return "", fmt.Errorf("both 'main' and 'master' branches exist")
	} else if errMain == nil {
		return "main", nil
	} else if errMaster == nil {
		return "master", nil
	} else {
		return "", fmt.Errorf("neither 'main' nor 'master' branch exists")
	}
}

func (r *RealGit) GetDiff() (string, error) {
	baseBranch, err := r.GetBaseBranchName()
	if err != nil {
		return "", fmt.Errorf("error determining base branch: %v", err)
	}
	out, diffErr := r.shell.RunCommandInDir(r.dir, "git", "diff", baseBranch, "--cached")
	if diffErr != nil {
		return "", err
	} else {
		diff := out
		names, err := r.shell.RunCommandInDir(r.dir, "git", "diff", baseBranch, "--cached", "--name-status")
		if err != nil {
			log.Fatalf("Can't run get a files status diff: '%v'", err)
		}
		stat, err := r.shell.RunCommandInDir(r.dir, "git", "diff", baseBranch, "--cached", "--stat")
		if err != nil {
			log.Fatalf("Can't run get a stat diff: '%v'", err)
		}
		return NewSummary(diff, stat, names).Render(), nil
	}
}

func (r *RealGit) GetCurrentDiff() (string, error) {
	unstaged, unErr := r.shell.RunCommandInDir(r.dir, "git", "diff")
	if unErr != nil {
		return "", unErr
	}
	staged, stErr := r.shell.RunCommandInDir(r.dir, "git", "diff", "--cached")
	if stErr != nil {
		return "", stErr
	}
	diff := unstaged + "\n" + staged
	return firstNLines(firstNChars(diff, 120*400), 400), nil
}

func (r *RealGit) GetCurrentCommitMessage() (string, error) {
	out, err := r.shell.RunCommandInDir(r.dir, "git", "log", "-1", "--pretty=%B")
	if err != nil {
		return "", err
	}
	commitMessage := out
	commitMessage = strings.TrimSpace(commitMessage)
	return commitMessage, nil
}

// This method returns a unique list of remote urls
func (r *RealGit) Remotes() ([]string, error) {
	out, err := r.shell.RunCommandInDir(r.dir, "git", "remote", "-v")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	seen := make(map[string]struct{})
	var result []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.Contains(line, "(fetch)") {
			continue
		}
		url := fields[1]
		if _, exists := seen[url]; !exists {
			seen[url] = struct{}{}
			result = append(result, url)
		}
	}
	return result, nil
}

func (r *RealGit) Installed() (bool, error) {
	_, err := r.shell.RunCommandInDir(r.dir, "git", "--version")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *RealGit) Root() (string, error) {
	out, err := r.shell.RunCommandInDir(r.dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		log.Printf("Can't find git root directory, '%v'", err)
		return out, err
	}
	return strings.TrimSpace(out), nil
}

func (r *RealGit) AddAll() error {
	_, err := r.shell.RunCommandInDir(r.dir, "git", "add", "--all")
	if err != nil {
		return fmt.Errorf("error adding all changes: %w", err)
	}
	return nil
}

func (r *RealGit) Checkout(branch string) error {
	_, err := r.shell.RunCommandInDir(r.dir, "git", "checkout", "-b", branch)
	if err != nil {
		return fmt.Errorf("error checking out branch %s: %w", branch, err)
	}
	return nil
}
