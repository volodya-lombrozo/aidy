package git

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/executor"
)

type real struct {
	dir   string
	shell executor.Executor
}

func NewGit(shell executor.Executor, dir ...string) (Git, error) {
	return NewGitFallback(shell, os.Getwd, dir...)
}

func NewGitFallback(shell executor.Executor, fallback func() (string, error), dir ...string) (Git, error) {
	var directory string
	if len(dir) > 0 && dir[0] != "" {
		directory = dir[0]
	} else {
		var err error
		directory, err = fallback()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}
	return &real{dir: directory, shell: shell}, nil
}

func (r *real) Run(arg ...string) (string, error) {
	out, err := r.shell.RunCommandInDir(r.dir, "git", arg...)
	if err != nil {
		return out, fmt.Errorf("error running git command: %w", err)
	}
	return out, nil
}

func (r *real) Amend(message string) error {
	_, err := r.Run("commit", "--amend", "-m", message)
	if err != nil {
		return fmt.Errorf("error amending commit: %w", err)
	}
	return nil
}

func (r *real) Reset(ref string) error {
	_, err := r.Run("reset", "--soft", ref)
	if err != nil {
		return fmt.Errorf("error resetting to %s: %w", ref, err)
	}
	return nil
}

func (r *real) Append() error {
	_, err := r.Run("add", "--all")
	if err != nil {
		return fmt.Errorf("error adding changes: %w", err)
	}
	_, err = r.Run("commit", "--amend", "--no-edit")
	if err != nil {
		return fmt.Errorf("error amending commit: %w", err)
	}
	return nil
}

func (r *real) CurrentBranch() (string, error) {
	branchName, err := r.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	branchName = strings.TrimRight(branchName, "\r\n")
	return branchName, nil
}

func (r *real) BaseBranch() (string, error) {
	_, errMain := r.Run("show-ref", "--verify", "--quiet", "refs/heads/main")
	_, errMaster := r.Run("show-ref", "--verify", "--quiet", "refs/heads/master")
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

func (r *real) Diff() (string, error) {
	base, err := r.BaseBranch()
	if err != nil {
		return "", fmt.Errorf("error determining base branch: %v", err)
	}
	out, diffErr := r.Run("diff", base, "--cached")
	if diffErr != nil {
		return "", err
	} else {
		diff := out
		names, err := r.Run("diff", base, "--cached", "--name-status")
		if err != nil {
			log.Fatalf("Can't run get a files status diff: '%v'", err)
		}
		stat, err := r.Run("diff", base, "--cached", "--stat")
		if err != nil {
			log.Fatalf("Can't run get a stat diff: '%v'", err)
		}
		return NewSummary(diff, stat, names).Render(), nil
	}
}

func (r *real) CurrentDiff() (string, error) {
	unstaged, unErr := r.Run("diff")
	if unErr != nil {
		return "", unErr
	}
	staged, stErr := r.Run("diff", "--cached")
	if stErr != nil {
		return "", stErr
	}
	diff := unstaged + "\n" + staged
	return firstNLines(firstNChars(diff, 120*400), 400), nil
}

func (r *real) CommitMessage() (string, error) {
	out, err := r.Run("log", "-1", "--pretty=%B")
	if err != nil {
		return "", err
	}
	commitMessage := out
	commitMessage = strings.TrimSpace(commitMessage)
	return commitMessage, nil
}

// This method returns a unique list of remote urls
func (r *real) Remotes() ([]string, error) {
	out, err := r.Run("remote", "-v")
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

func (r *real) Installed() (bool, error) {
	_, err := r.Run("--version")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *real) Root() (string, error) {
	out, err := r.Run("rev-parse", "--show-toplevel")
	if err != nil {
		log.Printf("Can't find git root directory, '%v'", err)
		return out, err
	}
	return strings.TrimSpace(out), nil
}

func (r *real) AddAll() error {
	_, err := r.Run("add", "--all")
	if err != nil {
		return fmt.Errorf("error adding all changes: %w", err)
	}
	return nil
}

func (r *real) Checkout(branch string) error {
	_, err := r.Run("checkout", "-b", branch)
	if err != nil {
		return fmt.Errorf("error checking out branch %s: %w", branch, err)
	}
	return nil
}

func (r *real) Tags(repo string) ([]string, error) {
	var err error
	if repo == "" {
		_, err = r.Run("fetch", "--tags")
	} else {
		_, err = r.Run("fetch", repo, "--tags")
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching tags: %w", err)
	}
	out, err := r.Run("tag")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return []string{}, nil
	}
	tags := strings.Split(strings.TrimSpace(out), "\n")
	return tags, nil
}

func (r *real) Log(since string) ([]string, error) {
	var args []string
	if since == "" {
		args = []string{"log", "--pretty=format:%s"}
	} else {
		args = []string{"log", fmt.Sprintf("%s..HEAD", since), "--pretty=format:%s"}
	}
	out, err := r.Run(args...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	return lines, nil
}
