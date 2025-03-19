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
