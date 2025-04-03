package executor

import (
    "bytes"
    "os/exec"
    "log"
    "strings"
)

type RealExecutor struct{}

func (r *RealExecutor) RunCommand(name string, args ...string) (string, error) {
    log.Printf("Executing command: %s %s", name, strings.Join(args, " "))
    cmd := exec.Command(name, args...)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return "", err
    }
    return out.String(), nil
}
