package git

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type RealGit struct {
	dir string
}

// NewRealGit creates a new RealGit instance. If no directory is provided, it uses the current working directory.
func NewRealGit(dir ...string) *RealGit {
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
	return &RealGit{dir: directory}
}

func (r *RealGit) GetBranchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = r.dir
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running git rev-parse: %v; stderr: %s", err, strings.TrimSpace(stderr.String()))
	}
	branchName := strings.TrimSpace(out.String())
	return branchName, nil
}

func (r *RealGit) GetBaseBranchName() (string, error) {
	log.Printf("Executing command to check base branch in directory: %s", r.dir)
	// Check if 'main' branch exists
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/main")
	cmd.Dir = r.dir
	errMain := cmd.Run()

	// Check if 'master' branch exists
	cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/master")
	cmd.Dir = r.dir
	errMaster := cmd.Run()

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
	// Determine the base branch name
	baseBranch, err := r.GetBaseBranchName()
	if err != nil {
		return "", fmt.Errorf("error determining base branch: %v", err)
	}
	cmd := exec.Command("git", "diff", baseBranch)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = r.dir
	err = cmd.Run()
	if err != nil {
		return "", err
	} else {
		diff := out.String()
		return diff, nil
	}
}

func (r *RealGit) GetCurrentCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = r.dir
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	commitMessage := out.String()
	commitMessage = strings.TrimSpace(commitMessage)
	return commitMessage, nil
}
