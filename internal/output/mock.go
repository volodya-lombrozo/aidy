package output

import "strings"

type Mock struct {
	captured    []string
	ReviewErr   error
	ReviewText  string
	Generations int
}

func NewMock() *Mock {
	return &Mock{}
}

func (m *Mock) Print(text Text) error {
	command, err := m.generate(text)
	if err != nil {
		return err
	}
	m.captured = append(m.captured, command)
	return nil
}

func (m *Mock) Review(text Text) (string, error) {
	current, err := m.generate(text)
	if err != nil {
		return "", err
	}
	m.captured = append(m.captured, current)
	if m.ReviewErr != nil {
		return "", m.ReviewErr
	}
	if m.ReviewText != "" {
		return m.ReviewText, nil
	}
	return current, nil
}

func (m *Mock) generate(text Text) (string, error) {
	var last string
	for range max(m.Generations, 1) {
		generated, err := text()
		if err != nil {
			return "", err
		}
		last = generated
	}
	return last, nil
}

func (m *Mock) Captured() string {
	if len(m.captured) == 0 {
		return ""
	}
	return strings.Join(m.captured, "\n")
}

func (m *Mock) Last() string {
	size := len(m.captured)
	if size < 1 {
		panic("we weren't able to capture anything")
	}
	return m.captured[size-1]
}

func (m *Mock) Replay(command string) error {
	m.captured = append(m.captured, command)
	return nil
}

type MockMemory struct {
	Commands []string
}

func NewMockMemory() *MockMemory {
	return &MockMemory{}
}

func (m *MockMemory) WithLast(command string) {
	m.Commands = append(m.Commands, command)
}
