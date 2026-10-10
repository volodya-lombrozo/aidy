package jira

import "fmt"

type offline struct{}

func NewOffline() *offline {
	return &offline{}
}

func (o *offline) IssueType(key string) (string, error) {
	return "", fmt.Errorf("jira connection is not configured, can't retrieve issue '%s'", key)
}
