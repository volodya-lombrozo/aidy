package git

import (
    "bytes"
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

func (r *RealGit) GetDiff(baseBranch string) (string, error) {
    cmd := exec.Command("git", "diff", baseBranch)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return "", err
    }
    diff := out.String()
    return diff, nil
}
