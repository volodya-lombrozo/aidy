package github

import "fmt"

type MockGithub struct {
	Error error
}

func NewMock() *MockGithub {
	return &MockGithub{Error: nil}
}

func (m *MockGithub) Description(number string) (string, error) {
	return fmt.Sprintf("mock description for issue '#%s'", number), m.Error
}

func (m *MockGithub) Labels() ([]string, error) {
	return []string{"bug", "documentation", "question"}, m.Error
}

func (m *MockGithub) Remotes() ([]string, error) {
	return []string{"volodya-lombrozo/aidy", "volodya-lombrozo/jtcop"}, m.Error
}
