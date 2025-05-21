package executor

type Executor interface {
	RunCommand(cmd string, args ...string) (string, error)

	RunCommandInDir(dir string, cmd string, args ...string) (string, error)

	RunInteractively(cmd string, args ...string) (string, error)
}
