package aidy

type Aidy interface {
	Release(interval string, repo string) error
	PrintConfig() error
	Commit() error
	Squash()
	PullRequest()
	Issue(task string)
	Heal() error
	Append()
	Clean()
	Diff() error
	StartIssue(number string) error
}
