package log

import "fmt"

type Mock struct {
	Messages []string
}

func NewMock() *Mock {
	return &Mock{Messages: []string{}}
}

func (m *Mock) Info(args ...any) {
	m.Messages = append(m.Messages, fmt.Sprintf("mock info: %v", args...))
}

func (m *Mock) Debug(args ...any) {
	m.Messages = append(m.Messages, fmt.Sprintf("mock dubug: %v", args...))
}

func (m *Mock) Warn(args ...any) {
	m.Messages = append(m.Messages, fmt.Sprintf("mock warn: %v", args...))
}
