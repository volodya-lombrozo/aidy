package log

import "fmt"

type Mock struct {
	Messages []string
}

func NewMock() *Mock {
	return &Mock{Messages: []string{}}
}

func (m *Mock) Info(msg string, args ...any) {
	m.Messages = append(m.Messages, fmt.Sprintf("mock info: %v", fmt.Sprintf(msg, args...)))
}

func (m *Mock) Debug(msg string, args ...any) {
	m.Messages = append(m.Messages, fmt.Sprintf("mock dubug: %v", fmt.Sprintf(msg, args...)))
}

func (m *Mock) Warn(msg string, args ...any) {
	m.Messages = append(m.Messages, fmt.Sprintf("mock warn: %v", fmt.Sprintf(msg, args...)))
}
