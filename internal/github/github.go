package github

type Github interface {
	Description(number string) (string, error)
	Labels() ([]string, error)
	Remotes() ([]string, error)
	PullRequestByBranch(branch string) (title string, body string, err error)
}
