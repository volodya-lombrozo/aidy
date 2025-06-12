package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/volodya-lombrozo/aidy/internal/log"
	"io"
	"net/http"
	"strings"
)

type DeepSeek struct {
	token   string
	url     string
	model   string
	summary bool
	log     log.Logger
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

func NewDeepSeek(apiKey string, summary bool) AI {
	return &DeepSeek{
		token:   apiKey,
		url:     "https://api.deepseek.com/chat/completions",
		model:   "deepseek-chat",
		summary: summary,
	}
}

func (d *DeepSeek) ReleaseNotes(changes string) (string, error) {
	prompt := fmt.Sprintf(ReleaseNotes, changes)
	return d.send("You are a helpful assistant generating GitHub release notes.", prompt, "")
}

func (d *DeepSeek) PrTitle(number, diff, issue, summary string) (string, error) {
	prompt := fmt.Sprintf(PrTitle, diff, issue, number, number)
	return d.send("You are a helpful assistant generating Git commit titles.", prompt, summary)
}

func (d *DeepSeek) PrBody(number string, diff string, issue string, summary string) (string, error) {
	prompt := fmt.Sprintf(PrBody, diff, issue, number)
	return d.send("You are a helpful assistant generating Git commit messages.", prompt, summary)
}

func (d *DeepSeek) IssueTitle(userInput string, summary string) (string, error) {
	prompt := fmt.Sprintf(IssueTitle, userInput)
	return d.send("You are a helpful assistant creating GitHub issue titles.", prompt, summary)
}

func (d *DeepSeek) IssueBody(input string, summary string) (string, error) {
	prompt := fmt.Sprintf(IssueBody, input)
	return d.send("You are a helpful assistant writing GitHub issue descriptions.", prompt, summary)
}

func (d *DeepSeek) IssueLabels(issue string, available []string) ([]string, error) {
	alllabels := strings.Join(available, ", ")
	prompt := fmt.Sprintf(Labels, issue, alllabels)
	resp, err := d.send("You are a helpful assistant assigning GitHub issue labels.", prompt, "")
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

func (d *DeepSeek) CommitMessage(number string, diff string) (string, error) {
	prompt := fmt.Sprintf(CommitMsg, diff, number, number)
	return d.send("You are a helpful assistant writing commit messages.", prompt, "")
}

func (d *DeepSeek) Summary(readme string) (string, error) {
	prompt := fmt.Sprintf(Summary, readme)
	return d.send("You are a helpful assistant writing project summaries.", prompt, "")
}

func (d *DeepSeek) SuggestBranch(descr string) (string, error) {
	prompt := fmt.Sprintf(BranchName, descr)
	return d.send("You are a helpful assistant suggesting branch names.", prompt, "")
}

func (d *DeepSeek) send(system string, user string, summary string) (string, error) {
	content := user
	if d.summary {
		content = AppendSummary(content, summary)
	}
	content = TrimPrompt(content)
	body := chatRequest{
		Model: d.model,
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
	req, err := http.NewRequest("POST", d.url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to deepseek api: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			d.log.Error("error closing response body: %v", err)
		}
	}()
	if resp.StatusCode != 200 {
		content, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", content)
	}
	var parsed chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}
