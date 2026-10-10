package jira

type Jira interface {
	IssueType(key string) (string, error)
}
