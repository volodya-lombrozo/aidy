package ai

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type openClient interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

type OpenAI struct {
	client      openClient
	model       string
	temperature float32
	summary     bool
}

func NewOpenAI(token, model string, temperature float32, summary bool) *OpenAI {
	return NewOpenAIWithClient(openai.NewClient(token), model, temperature, summary)
}

func NewOpenAIWithClient(client openClient, model string, temperature float32, summary bool) *OpenAI {
	return &OpenAI{
		client:      client,
		model:       model,
		temperature: temperature,
		summary:     summary,
	}
}

func (o *OpenAI) ReleaseNotes(changes string) (string, error) {
	prompt := fmt.Sprintf(ReleaseNotes, changes)
	return o.send(prompt, "")
}

func (o *OpenAI) PrTitle(number, diff, issue, summary string) (string, error) {
	prompt := fmt.Sprintf(PrTitle, diff, issue, number, number)
	return o.send(prompt, summary)
}

func (o *OpenAI) PrBody(diff, issue, summary string) (string, error) {
	prompt := fmt.Sprintf(PrBody, diff, issue)
	return o.send(prompt, summary)
}

func (o *OpenAI) IssueTitle(input, summary string) (string, error) {
	prompt := fmt.Sprintf(IssueTitle, input)
	return o.send(prompt, summary)
}

func (o *OpenAI) IssueBody(input string, summary string) (string, error) {
	prompt := fmt.Sprintf(IssueBody, input)
	return o.send(prompt, summary)
}

func (o *OpenAI) CommitMessage(number, diff string) (string, error) {
	prompt := fmt.Sprintf(CommitMsg, diff, number, number)
	return o.send(prompt, "")
}

func (o *OpenAI) IssueLabels(issue string, available []string) ([]string, error) {
	alllabels := strings.Join(available, ", ")
	prompt := fmt.Sprintf(Labels, issue, alllabels)
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

func (o *OpenAI) Summary(readme string) (string, error) {
	prompt := fmt.Sprintf(Summary, readme)
	return o.send(prompt, "")
}

func (o *OpenAI) SuggestBranch(descr string) (string, error) {
	prompt := fmt.Sprintf(BranchName, descr)
	return o.send(prompt, "")
}

// send sends a prompt to the OpenAI API and returns the response.
// Parameters:
// - prompt: The prompt to send.
// - summary: The project summary to append to the prompt (if applicable).
func (o *OpenAI) send(prompt, summary string) (string, error) {
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
