package output

import "strings"

type Mock struct {
	captured   []string
	ReviewErr  error
	ReviewText string
}

func NewMock() *Mock {
	return &Mock{}
}

func (m *Mock) Print(command string) error {
	m.captured = append(m.captured, command)
	return nil
}

func (m *Mock) Review(text string) (string, error) {
	m.captured = append(m.captured, text)
	if m.ReviewErr != nil {
		return "", m.ReviewErr
	}
	if m.ReviewText != "" {
		return m.ReviewText, nil
	}
	return text, nil
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
