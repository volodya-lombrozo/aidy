package executor

import "strings"

type MockExecutor struct {
	Output   string
	Err      error
	Commands []string
}

func NewMock() *MockExecutor {
	return &MockExecutor{}
}

func (m *MockExecutor) RunInteractively(cmd string, args ...string) (string, error) {
	command := cmd + " " + strings.Join(args, " ")
	m.Commands = append(m.Commands, command)
	return m.Output, m.Err
}

func (m *MockExecutor) RunCommand(cmd string, args ...string) (string, error) {
	command := cmd + " " + strings.Join(args, " ")
	m.Commands = append(m.Commands, command)
	return m.Output, m.Err
}

func (m *MockExecutor) RunCommandInDir(dir string, cmd string, args ...string) (string, error) {
	command := "cd " + dir + " && " + cmd + " " + strings.Join(args, " ")
	m.Commands = append(m.Commands, command)
	return m.Output, m.Err
}
