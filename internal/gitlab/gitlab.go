package gitlab

type Gitlab interface {
	MergeRequestByBranch(branch string) (title string, body string, err error)
}
