package executor

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type RealExecutor struct{}

const maxLogLength = 120

func (r *RealExecutor) RunCommand(name string, args ...string) (string, error) {
	logShortf("Execute: \"%s %s\"", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (r *RealExecutor) RunCommandInDir(dir string, name string, args ...string) (string, error) {
	logShortf("Execute (%s): \"%s %s\"", dir, name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error: %v; stderr: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out.String(), nil
}

func logShortf(format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	if len(msg) > maxLogLength {
		msg = msg[:maxLogLength] + "…"
	}
	msg = strings.ReplaceAll(msg, "\n", " ")
	log.Print(msg)
}
