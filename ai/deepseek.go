package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type DeepSeekAI struct {
	APIKey  string
	APIURL  string // e.g., "https://api.deepseek.com/v1/chat/completions"
	Model   string // e.g., "deepseek-chat" or similar
	summary bool
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

func NewDeepSeekAI(apiKey string, summary bool) *DeepSeekAI {
	return &DeepSeekAI{
		APIKey:  apiKey,
		APIURL:  "https://api.deepseek.com/chat/completions",
		Model:   "deepseek-chat",
		summary: summary,
	}
}

func (d *DeepSeekAI) PrTitle(number, diff, issue, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateTitlePrompt, diff, issue, number, number)
	return d.sendPrompt("You are a helpful assistant generating Git commit titles.", prompt, summary)
}

func (d *DeepSeekAI) PrBody(number string, diff string, issue string, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateBodyPrompt, diff, issue, number)
	return d.sendPrompt("You are a helpful assistant generating Git commit messages.", prompt, summary)
}

func (d *DeepSeekAI) IssueTitle(userInput string, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateIssueTitlePrompt, userInput)
	return d.sendPrompt("You are a helpful assistant creating GitHub issue titles.", prompt, summary)
}

func (d *DeepSeekAI) IssueBody(input string, summary string) (string, error) {
	prompt := fmt.Sprintf(GenerateIssueBodyPrompt, input)
	return d.sendPrompt("You are a helpful assistant writing GitHub issue descriptions.", prompt, summary)
}

func (d *DeepSeekAI) IssueLabels(issue string, available []string) ([]string, error) {
	alllabels := strings.Join(available, ", ")
	prompt := fmt.Sprintf(GenerateLabelsPrompt, issue, alllabels)
	resp, err := d.sendPrompt("You are a helpful assistant assigning GitHub issue labels.", prompt, "")
	if err != nil {
		return nil, err
	}
	var res []string
	for _, label := range available {
		if strings.Contains(resp, label) {
			res = append(res, label)
		}
	}
	return res, nil
}

func (d *DeepSeekAI) CommitMessage(number string, diff string) (string, error) {
	prompt := fmt.Sprintf(GenerateCommitPrompt, diff, number, number)
	return d.sendPrompt("You are a helpful assistant writing commit messages.", prompt, "")
}

func (d *DeepSeekAI) Summary(readme string) (string, error) {
	prompt := fmt.Sprintf(SummaryPrompt, readme)
	return d.sendPrompt("You are a helpful assistant writing project summaries.", prompt, "")
}

func (d *DeepSeekAI) sendPrompt(system string, user string, summary string) (string, error) {
	content := user
	if d.summary {
		content = AppendSummary(content, summary)
	}
	content = TrimPrompt(content)
	body := chatRequest{
		Model: d.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: content},
		},
		Stream: false,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", d.APIURL, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.APIKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != 200 {
		content, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", content)
	}

	var parsed chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	if len(parsed.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}
