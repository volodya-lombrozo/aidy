package git

type Git interface {
    GetBranchName() (string, error)
}
