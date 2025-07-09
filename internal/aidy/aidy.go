package aidy

type Aidy interface {
	Release(interval, repo string) error
	PrintConfig() error
	Commit(issue bool) error
	Squash(issue bool)
	PullRequest(fixes bool) error
	Issue(task string) error
	Heal() error
	Append()
	Clean()
	Diff() error
	StartIssue(number string) error
}
