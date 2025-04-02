package executor

type Executor interface {
    RunCommand(name string, args ...string) (string, error)
}
