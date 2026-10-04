package cache

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/git"
)

type LastCommand interface {
	Last() (string, error)
	WithLast(command string)
}

type lastCommand struct {
	gs git.Git
}

func NewLastCommand(gs git.Git) LastCommand {
	return &lastCommand{gs: gs}
}

func (l *lastCommand) Last() (string, error) {
	path, err := l.path()
	if err != nil {
		return "", err
	}
	command, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("can't read the last command from '%s': %w", path, err)
	}
	return string(command), nil
}

func (l *lastCommand) WithLast(command string) {
	path, err := l.path()
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, []byte(command), 0644); err != nil {
		panic(fmt.Errorf("can't save the last command to '%s', because '%v'", path, err))
	}
}

func (l *lastCommand) path() (string, error) {
	out, err := l.gs.Run("rev-parse", "--path-format=absolute", "--git-path", "aidy-last")
	if err != nil {
		return "", fmt.Errorf("can't find where to keep the last command: %w", err)
	}
	return strings.TrimSpace(out), nil
}

type mockLastCommand struct {
	command string
}

func NewMockLastCommand() LastCommand {
	return &mockLastCommand{}
}

func (m *mockLastCommand) Last() (string, error) {
	return m.command, nil
}

func (m *mockLastCommand) WithLast(command string) {
	m.command = command
}
