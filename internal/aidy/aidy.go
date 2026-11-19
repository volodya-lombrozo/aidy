package aidy

type Aidy interface {
	Release(interval, repo string, notes bool) error
	PrintConfig() error
	Commit(issue bool) error
	Squash(issue bool)
	PullRequest(fixes bool, target string, duplicate bool, source string) error
	MergeRequest(fixes bool, target string, duplicate bool, source string) error
	Issue(task string) error
	Append()
	Clean()
	StartIssue(number string) error
	Last() error
}
