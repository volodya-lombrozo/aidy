package git

type Git interface {
	GetBranchName() (string, error)
	GetDiff() (string, error)
	GetCurrentDiff() (string, error)
	GetBaseBranchName() (string, error)
	GetCurrentCommitMessage() (string, error)
	AppendToCommit() error
	CommitChanges(messages ...string) error
	Remotes() ([]string, error)
	Installed() (bool, error)
	Root() (string, error)
	Reset(ref string) error
	AddAll() error
	Amend(message string) error
}
