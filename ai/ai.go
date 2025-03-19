package ai

type AI interface {
    GenerateTitle(branchName string) (string, error)
    GenerateBody(branchName string) (string, error)
}
