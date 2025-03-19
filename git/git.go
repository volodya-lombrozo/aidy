package git

type Git interface {
    GetBranchName() (string, error)
    GetDiff(baseBranch string) (string, error)
}
