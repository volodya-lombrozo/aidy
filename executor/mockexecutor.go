package executor

import "strings"

type MockExecutor struct {
    Output string
    Err    error
    Commands []string
}

func (m *MockExecutor) RunCommand(name string, args ...string) (string, error) {
    command := name + " " + strings.Join(args, " ")
    m.Commands = append(m.Commands, command)
    return m.Output, m.Err
}

