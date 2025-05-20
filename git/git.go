package git

// This file defines the Git interface and its methods for interacting with git repositories.
// In most cases only the Run method is used, which executes git commands.
// All other methods are wrappers around Run.
// Don't add new methods to this interface.
// If you need to add a new method, consider if it can be implemented using Run.
type Git interface {
	Run(args ...string) (string, error)

	Installed() (bool, error)
	CurrentBranch() (string, error)
	BaseBranch() (string, error)
	Diff() (string, error)
	CurrentDiff() (string, error)
	CommitMessage() (string, error)
	Append() error
	Remotes() ([]string, error)
	Root() (string, error)
	Reset(ref string) error
	AddAll() error
	Amend(message string) error
	Checkout(branch string) error
	Tags(repo string) ([]string, error)
	Log(since string) ([]string, error)
}
