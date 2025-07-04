package aidy

type Aidy interface {
	Release(interval string, repo string) error
	PrintConfig() error
	Commit() error
	Squash()
	PullRequest(fixes bool) error
	Issue(task string) error
	Heal() error
	Append()
	Clean()
	Diff() error
	StartIssue(number string) error
}
