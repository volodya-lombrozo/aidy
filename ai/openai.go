package ai

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"strings"
)

type MyOpenAI struct {
	client      *openai.Client
	model       string
	temperature float32
}

func NewOpenAI(apiKey, model string, temperature float32) *MyOpenAI {
	client := openai.NewClient(apiKey)
	return &MyOpenAI{
		client:      client,
		model:       model,
		temperature: temperature,
	}
}

func (o *MyOpenAI) GenerateTitle(branchName, diff string, issue string) (string, error) {
	issueNumber := extractIssueNumber(branchName)
	prompt := fmt.Sprintf(GenerateTitlePrompt, diff, issue, issueNumber, issueNumber)
	return o.generateText(prompt)
}

func (o *MyOpenAI) GenerateBody(branchName, diff string, issue string) (string, error) {
	issueNumber := extractIssueNumber(branchName)
	prompt := fmt.Sprintf(GenerateBodyPrompt, diff, issue, issueNumber)
	return o.generateText(prompt)
}

func (o *MyOpenAI) GenerateIssueTitle(userInput string) (string, error) {
	prompt := fmt.Sprintf(GenerateIssueTitlePrompt, userInput)
	return o.generateText(prompt)
}

func (o *MyOpenAI) GenerateIssueBody(userInput string) (string, error) {
	prompt := fmt.Sprintf(GenerateIssueBodyPrompt, userInput)
	return o.generateText(prompt)
}

func (o *MyOpenAI) GenerateCommitMessage(branchName, diff string) (string, error) {
	issueNumber := extractIssueNumber(branchName)
	prompt := fmt.Sprintf(GenerateCommitPrompt, diff, issueNumber, issueNumber)
	return o.generateText(prompt)
}

func (o *MyOpenAI) GenerateIssueLabels(issue string, available []string) ([]string, error) {
    alllabels := strings.Join(available, ", ")
   	prompt := fmt.Sprintf(GenerateLabelsPrompt, issue, alllabels)
    out, err := o.generateText(prompt)
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

func (o *MyOpenAI) generateText(prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: prompt,
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
