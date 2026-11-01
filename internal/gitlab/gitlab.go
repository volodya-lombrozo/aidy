package gitlab

type Gitlab interface {
	Description(number string, repository string) (string, error)
	MergeRequestByBranch(branch string) (title string, body string, err error)
}
