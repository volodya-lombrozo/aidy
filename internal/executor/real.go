package executor

import (
	"bytes"
	"fmt"
	"github.com/volodya-lombrozo/aidy/internal/log"
	"os"
	"os/exec"
	"strings"
)

type RealExecutor struct {
	out *os.File
	in  *os.File
	err *os.File
	log log.Logger
}

func NewReal() Executor {
	return &RealExecutor{
		out: os.Stdout,
		in:  os.Stdin,
		err: os.Stderr,
		log: log.Get(),
	}
}

func (r *RealExecutor) RunInteractively(cmd string, args ...string) (string, error) {
	r.log.Debug("execute: \"%s\" with args(%d): \"%v\"", cmd, len(args), args)
	command := exec.Command(cmd, args...)
	command.Stdout = r.out
	command.Stderr = r.err
	command.Stdin = r.in
	err := command.Run()
	if err != nil {
		return "", err
	}
	return "unimplemented", nil
}

func (r *RealExecutor) RunCommand(name string, args ...string) (string, error) {
	r.log.Debug("execute: \"%s %s\"", name, strings.Join(args, " "))
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
	r.log.Debug("execute (%s): \"%s %s\"", dir, name, strings.Join(args, " "))
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
