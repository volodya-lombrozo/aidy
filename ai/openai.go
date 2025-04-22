package ai

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type MyOpenAI struct {
	client      *openai.Client
	model       string
	temperature float32
	summary     bool
}

func NewOpenAI(apiKey, model string, temperature float32, summary bool) *MyOpenAI {
	client := openai.NewClient(apiKey)
	return &MyOpenAI{
		client:      client,
		model:       model,
		temperature: temperature,
		summary:     summary,
	}
}

func (o *MyOpenAI) PrTitle(branchName, diff string, issue string, summary string) (string, error) {
	issueNumber := extractIssueNumber(branchName)
	prompt := fmt.Sprintf(GenerateTitlePrompt, diff, issue, issueNumber, issueNumber)
	return o.generateText(prompt, summary)
}

func (o *MyOpenAI) PrBody(branchName, diff string, issue string, summary string) (string, error) {
	issueNumber := extractIssueNumber(branchName)
	prompt := fmt.Sprintf(GenerateBodyPrompt, diff, issue, issueNumber)
	return o.generateText(prompt, summary)
}

func (o *MyOpenAI) IssueTitle(userInput string, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateIssueTitlePrompt, userInput)
	return o.generateText(prompt, summary)
}

func (o *MyOpenAI) IssueBody(userInput string, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateIssueBodyPrompt, userInput)
	return o.generateText(prompt, summary)
}

func (o *MyOpenAI) CommitMessage(branchName, diff string) (string, error) {
	issueNumber := extractIssueNumber(branchName)
	prompt := fmt.Sprintf(GenerateCommitPrompt, diff, issueNumber, issueNumber)
	return o.generateText(prompt, "")
}

func (o *MyOpenAI) IssueLabels(issue string, available []string) ([]string, error) {
	alllabels := strings.Join(available, ", ")
	prompt := fmt.Sprintf(GenerateLabelsPrompt, issue, alllabels)
	out, err := o.generateText(prompt, "")
	if err != nil {
		return nil, err
	}
	var res []string
	for _, label := range available {
		if strings.Contains(out, label) {
			res = append(res, label)
		}
	}
	return res, nil
}

func (o *MyOpenAI) Summary(readme string) (string, error) {
	prompt := fmt.Sprintf(SummaryPrompt, readme)
	return o.generateText(prompt, "")
}

func (o *MyOpenAI) generateText(prompt, summary string) (string, error) {
	content := prompt
	if o.summary {
		content = AppendSummary(content, summary)
	}
	content = TrimPrompt(content)
	req := openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: content,
			},
		},
		Temperature: o.temperature,
	}
	resp, err := o.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no text generated")
}

func extractIssueNumber(branchName string) string {
	// Assuming the branch name format is "<issue-number>_<description>"
	parts := strings.Split(branchName, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}
