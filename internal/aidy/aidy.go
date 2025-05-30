package aidy

type Aidy interface {
	Release(interval string, repo string) error
	PrintConfig() error
	Commit()
	Squash()
	PullRequest()
	Issue(task string)
	Heal()
	Append()
	Clean()
	Diff() error
	StartIssue(number string) error
}
