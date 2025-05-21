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

func (o *MyOpenAI) ReleaseNotes(changes string) (string, error) {
	prompt := fmt.Sprintf(ReleaseNotesPrompt, changes)
	return o.send(prompt, "")
}

func (o *MyOpenAI) PrTitle(number, diff, issue, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateTitlePrompt, diff, issue, number, number)
	return o.send(prompt, summary)
}

func (o *MyOpenAI) PrBody(number, diff, issue, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateBodyPrompt, diff, issue, number)
	return o.send(prompt, summary)
}

func (o *MyOpenAI) IssueTitle(input, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateIssueTitlePrompt, input)
	return o.send(prompt, summary)
}

func (o *MyOpenAI) IssueBody(input string, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateIssueBodyPrompt, input)
	return o.send(prompt, summary)
}

func (o *MyOpenAI) CommitMessage(number, diff string) (string, error) {
	prompt := fmt.Sprintf(GenerateCommitPrompt, diff, number, number)
	return o.send(prompt, "")
}

func (o *MyOpenAI) IssueLabels(issue string, available []string) ([]string, error) {
	alllabels := strings.Join(available, ", ")
	prompt := fmt.Sprintf(GenerateLabelsPrompt, issue, alllabels)
	out, err := o.send(prompt, "")
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
	return o.send(prompt, "")
}

func (o *MyOpenAI) SuggestBranch(descr string) (string, error) {
	prompt := fmt.Sprintf(SuggestBranchPrompt, descr)
	return o.send(prompt, "")
}

// send sends a prompt to the OpenAI API and returns the response.
// Parameters:
// - prompt: The prompt to send.
// - summary: The project summary to append to the prompt (if applicable).
func (o *MyOpenAI) send(prompt, summary string) (string, error) {
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
