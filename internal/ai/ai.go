package ai

import "fmt"

type AI interface {
	PrTitle(number string, diff string, issue string, summary string) (string, error)
	PrBody(number string, diff string, issue string, summary string) (string, error)
	IssueTitle(input string, summary string) (string, error)
	IssueBody(input string, summary string) (string, error)
	IssueLabels(issue string, available []string) ([]string, error)
	CommitMessage(number string, diff string) (string, error)
	Summary(readme string) (string, error)
	SuggestBranch(descr string) (string, error)
	ReleaseNotes(changes string) (string, error)
}

func TrimPrompt(prompt string) string {
	limit := 120 * 400
	runes := []rune(prompt)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return prompt
}

func AppendSummary(prompt, summary string) string {
	if summary == "" {
		return prompt
	}
	appendix := fmt.Sprintf("\nThis is the project summary for which you do it:\n<summary>\n%s\n</summary>\n", summary)
	return prompt + appendix
}
