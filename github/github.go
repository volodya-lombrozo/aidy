package github

type Github interface {
	IssueDescription(number string) string
    Labels() []string
}
