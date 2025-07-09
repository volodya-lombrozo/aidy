package ai

import "fmt"

type AI interface {
	PrTitle(number, diff, issue, summary string) (string, error)
	PrBody(diff, issue, summary string) (string, error)
	IssueTitle(input, summary string) (string, error)
	IssueBody(input, summary string) (string, error)
	IssueLabels(issue string, available []string) ([]string, error)
	CommitMessage(number, diff, descr string) (string, error)
	Summary(readme string) (string, error)
	SuggestBranch(descr string) (string, error)
	ReleaseNotes(changes string) (string, error)
}

func trimPrompt(prompt string) string {
	limit := 120 * 400
	runes := []rune(prompt)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return prompt
}

func appendSummary(prompt, summary string) string {
	if summary == "" {
		return prompt
	}
	appendix := fmt.Sprintf("\nThis is the project summary for which you do it:\n<summary>\n%s\n</summary>\n", summary)
	return prompt + appendix
}

func appendIssue(prompt, description string) string {
	if description == "" {
		return prompt
	}
	appendix := fmt.Sprintf("\nThis is the issue description for which you do it:\n<issue>\n%s\n</issue>\n", description)
	return prompt + appendix
}
