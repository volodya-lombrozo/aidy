package output

type Output interface {
	Print(command string) error
}
