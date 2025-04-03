package git

type Git interface {
    GetBranchName() (string, error)
    GetDiff() (string, error)
    GetBaseBranchName() (string, error)
    GetCurrentCommitMessage() (string, error)
}
