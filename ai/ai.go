package ai

type AI interface {
	GenerateTitle(branchName string, diff string, issue string) (string, error)
	GenerateBody(branchName string, diff string, issue string) (string, error)
	GenerateIssueTitle(userInput string) (string, error)
	GenerateIssueBody(userInput string) (string, error)
	GenerateIssueLabels(issue string, available []string) ([]string, error)
	GenerateCommitMessage(branchName string, diff string) (string, error)
}

func TrimPrompt(prompt string) string {
    limit := 120 * 400
	runes := []rune(prompt)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return prompt
}
