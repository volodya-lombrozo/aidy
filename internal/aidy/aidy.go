package aidy

type Aidy interface {
	Release(interval string, repo string) error
	PrintConfig()
	Commit()
	Squash()
	PullRequest()
	Issue(task string)
	Heal()
	Append()
	Clean()
	Diff()
	StartIssue(number string) error
}
