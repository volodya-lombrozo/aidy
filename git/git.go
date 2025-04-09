package git

type Git interface {
	GetBranchName() (string, error)
	GetDiff() (string, error)
	GetBaseBranchName() (string, error)
	GetCurrentCommitMessage() (string, error)
	AppendToCommit() error
	CommitChanges() error
	GetAllRemoteURLs() ([]string, error)
}
