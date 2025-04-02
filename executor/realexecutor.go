package executor

import (
    "bytes"
    "os/exec"
)

type RealExecutor struct{}

func (r *RealExecutor) RunCommand(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return "", err
    }
    return out.String(), nil
}
