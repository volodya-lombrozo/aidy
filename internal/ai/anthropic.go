package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/log"
)

const anthropicDefaultModel = "claude-sonnet-4-6"
const anthropicVersion = "2023-06-01"

type Anthropic struct {
	token   string
	url     string
	model   string
	summary bool
	log     log.Logger
}

type anthropicRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system"`
	Messages  []chatMessage `json:"messages"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
}

func NewAnthropic(token, model string, summary bool) AI {
	if model == "" {
		model = anthropicDefaultModel
	}
	return &Anthropic{
		token:   token,
		url:     "https://api.anthropic.com/v1/messages",
		model:   model,
		summary: summary,
		log:     log.Default(),
	}
}

func (a *Anthropic) ReleaseNotes(changes string) (string, error) {
	prompt := fmt.Sprintf(ReleaseNotes, changes)
	return a.send("You are a helpful assistant generating GitHub release notes.", prompt, "")
}

func (a *Anthropic) PrTitle(number, diff, issue, summary string) (string, error) {
	prompt := fmt.Sprintf(PrTitle, diff, issue, number, number)
	return a.send("You are a helpful assistant generating Git commit titles.", prompt, summary)
}

func (a *Anthropic) PrBody(diff string, issue string, summary string) (string, error) {
	prompt := fmt.Sprintf(PrBody, diff, issue)
	return a.send("You are a helpful assistant generating Git commit messages.", prompt, summary)
}

func (a *Anthropic) IssueTitle(userInput string, summary string) (string, error) {
	prompt := fmt.Sprintf(IssueTitle, userInput)
	return a.send("You are a helpful assistant creating GitHub issue titles.", prompt, summary)
}

func (a *Anthropic) IssueBody(input string, summary string) (string, error) {
	prompt := fmt.Sprintf(IssueBody, input)
	return a.send("You are a helpful assistant writing GitHub issue descriptions.", prompt, summary)
}

func (a *Anthropic) IssueLabels(issue string, available []string) ([]string, error) {
	alllabels := strings.Join(available, ", ")
	prompt := fmt.Sprintf(Labels, issue, alllabels)
	resp, err := a.send("You are a helpful assistant assigning GitHub issue labels.", prompt, "")
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

func (a *Anthropic) CommitMessage(number, diff, descr string) (string, error) {
	prompt := appendIssue(fmt.Sprintf(CommitMsg, diff, number, number), descr)
	a.log.Debug("anthropic prompt: %q", prompt)
	return a.send("You are a helpful assistant writing commit messages.", prompt, "")
}

func (a *Anthropic) Summary(readme string) (string, error) {
	prompt := fmt.Sprintf(Summary, readme)
	return a.send("You are a helpful assistant writing project summaries.", prompt, "")
}

func (a *Anthropic) SuggestBranch(descr string) (string, error) {
	prompt := fmt.Sprintf(BranchName, descr)
	return a.send("You are a helpful assistant suggesting branch names.", prompt, "")
}

func (a *Anthropic) send(system, user, summary string) (string, error) {
	content := user
	if a.summary {
		content = appendSummary(content, summary)
	}
	content = trimPrompt(content)
	body := anthropicRequest{
		Model:     a.model,
		MaxTokens: 1024,
		System:    system,
		Messages:  []chatMessage{{Role: "user", Content: content}},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", a.url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.token)
	req.Header.Set("anthropic-version", anthropicVersion)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to anthropic api: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.log.Error("error closing response body: %v", err)
		}
	}()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", body)
	}
	var parsed anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}
	if len(parsed.Content) == 0 {
		return "", errors.New("no content in response")
	}
	return strings.TrimSpace(parsed.Content[0].Text), nil
}
