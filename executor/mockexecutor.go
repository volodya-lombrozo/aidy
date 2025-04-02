package executor

type MockExecutor struct {
    Output string
    Err    error
}

func (m *MockExecutor) RunCommand(name string, args ...string) (string, error) {
    return m.Output, m.Err
}
