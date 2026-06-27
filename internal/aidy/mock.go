package aidy

import (
	"errors"
	"fmt"
)

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

func (m *Mock) Commit(issue bool) error {
	m.logs = append(m.logs, "Commit called")
	return nil
}

func (m *Mock) Squash(issue bool) {
	m.logs = append(m.logs, "Squash called")
}

func (m *Mock) PullRequest(fixes bool) error {
	m.logs = append(m.logs, "PullRequest called")
	return nil
}

func (m *Mock) MergeRequest(fixes bool) error {
	m.logs = append(m.logs, "MergeRequest called")
	return nil
}

func (m *Mock) Issue(task string) error {
	m.logs = append(m.logs, fmt.Sprintf("Issue called with task: %s", task))
	return nil
}

func (m *Mock) Heal() error {
	m.logs = append(m.logs, "Heal called")
	return nil
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

type FailingMock struct{}

func NewFailingMock() *FailingMock {
	return &FailingMock{}
}

func (f *FailingMock) Release(interval string, repo string) error   { return errors.New("error") }
func (f *FailingMock) PrintConfig() error                           { return errors.New("error") }
func (f *FailingMock) Commit(issue bool) error                      { return errors.New("error") }
func (f *FailingMock) Squash(issue bool)                            {}
func (f *FailingMock) PullRequest(fixes bool) error                 { return errors.New("error") }
func (f *FailingMock) MergeRequest(fixes bool) error                { return errors.New("error") }
func (f *FailingMock) Issue(task string) error                      { return errors.New("error") }
func (f *FailingMock) Heal() error                                   { return errors.New("error") }
func (f *FailingMock) Append()                                       {}
func (f *FailingMock) Clean()                                        {}
func (f *FailingMock) Diff() error                                   { return errors.New("error") }
func (f *FailingMock) StartIssue(number string) error               { return errors.New("error") }
