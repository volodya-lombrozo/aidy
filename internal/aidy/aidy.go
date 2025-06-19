package aidy

// @todo #185:15min Lack of documentation
//
//	This type lacking of documentation,
//	espesially on error cases, functions params
//	and etc.
type Aidy interface {
	Release(interval string, repo string) error
	PrintConfig() error
	Commit() error
	Squash()
	PullRequest() error
	Issue(task string) error
	Heal() error
	Append()
	Clean()
	Diff() error
	StartIssue(number string) error
	Repeat(path string) int
}
