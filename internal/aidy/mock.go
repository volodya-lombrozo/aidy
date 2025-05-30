package aidy

import "fmt"

type Mock struct {
	logs []string
}

func NewMock() *Mock {
	return &Mock{logs: []string{}}
}

func (m *Mock) Release(interval string, repo string) error {
	m.logs = append(m.logs, fmt.Sprintf("Release called with interval: %s, repo: %s", interval, repo))
	return nil
}

func (m *Mock) PrintConfig() error {
	m.logs = append(m.logs, "PrintConfig called")
	return nil
}

func (m *Mock) Commit() {
	m.logs = append(m.logs, "Commit called")
}

func (m *Mock) Squash() {
	m.logs = append(m.logs, "Squash called")
}

func (m *Mock) PullRequest() {
	m.logs = append(m.logs, "PullRequest called")
}

func (m *Mock) Issue(task string) {
	m.logs = append(m.logs, fmt.Sprintf("Issue called with task: %s", task))
}

func (m *Mock) Heal() {
	m.logs = append(m.logs, "Heal called")
}

func (m *Mock) Append() {
	m.logs = append(m.logs, "Append called")
}

func (m *Mock) Clean() {
	m.logs = append(m.logs, "Clean called")
}

func (m *Mock) Diff() error {
	m.logs = append(m.logs, "Diff called")
	return nil
}

func (m *Mock) StartIssue(number string) error {
	m.logs = append(m.logs, fmt.Sprintf("StartIssue called with number: %s", number))
	return nil
}

func (m *Mock) Logs() []string {
	return m.logs
}
