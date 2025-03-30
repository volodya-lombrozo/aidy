package git

import (
    "bytes"
    "fmt"
    "os/exec"
    "strings"
)

type RealGit struct{}

func (r *RealGit) GetBranchName() (string, error) {
    cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return "", err
    }
    branchName := strings.TrimSpace(out.String())
    return branchName, nil
}

func (r *RealGit) GetBaseBranchName() (string, error) {
    // Check if 'main' branch exists
    cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/main")
    errMain := cmd.Run()

    // Check if 'master' branch exists
    cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/master")
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
    err = cmd.Run()
    if err != nil {
      return "", err
    } else {
      diff := out.String()
      return diff, nil
    }
}
