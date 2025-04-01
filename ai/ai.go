package ai

type AI interface {
    GenerateTitle(branchName string, diff string) (string, error)
    GenerateBody(branchName string, diff string) (string, error)
    GenerateIssueTitle(userInput string) (string, error)
    GenerateIssueBody(userInput string) (string, error)
}
