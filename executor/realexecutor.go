package executor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type RealExecutor struct{}

const maxLogLength = 120

func NewRealExecutor() Executor {
	return &RealExecutor{}
}

func (r *RealExecutor) RunInteractively(cmd string, args ...string) (string, error) {
	logShortf("Execute: \"%s\" with args: \"%v\"", cmd, args)
	command := exec.Command(cmd, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	err := command.Run()
	if err != nil {
		return "", err
	}
	return "unimplemented", nil
}

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
