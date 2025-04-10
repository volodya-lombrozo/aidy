package git

type Git interface {
	GetBranchName() (string, error)
	GetDiff() (string, error)
	GetCurrentDiff() (string, error)
	GetBaseBranchName() (string, error)
	GetCurrentCommitMessage() (string, error)
	AppendToCommit() error
	CommitChanges(messages ...string) error
	GetAllRemoteURLs() ([]string, error)
}
