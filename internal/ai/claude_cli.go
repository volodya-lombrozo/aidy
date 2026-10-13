package ai

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/log"
)

const claudeCliDefaultModel = "sonnet"
const claudeCliBinary = "claude"

type ClaudeCli struct {
	binary   string
	model    string
	summary  bool
	language string
	log      log.Logger
}

func NewClaudeCli(model string, summary bool, language string) AI {
	if model == "" {
		model = claudeCliDefaultModel
	}
	return &ClaudeCli{
		binary:   claudeCliBinary,
		model:    model,
		summary:  summary,
		language: language,
		log:      log.Default(),
	}
}

func (c *ClaudeCli) ReleaseNotes(changes string) (string, error) {
	prompt := fmt.Sprintf(ReleaseNotes, trimInput(changes))
	return c.send("You are a helpful assistant generating GitHub release notes.", prompt, "")
}

func (c *ClaudeCli) PrTitle(number, diff, issue, summary string) (string, error) {
	prompt := fmt.Sprintf(PrTitle, trimInput(diff), issue, number, number)
	return c.send("You are a helpful assistant generating Git commit titles.", prompt, summary)
}

func (c *ClaudeCli) PrBody(diff string, issue string, summary string) (string, error) {
	prompt := fmt.Sprintf(PrBody, trimInput(diff), issue)
	return c.send("You are a helpful assistant generating Git commit messages.", prompt, summary)
}

func (c *ClaudeCli) IssueTitle(userInput string, summary string) (string, error) {
	prompt := fmt.Sprintf(IssueTitle, userInput)
	return c.send("You are a helpful assistant creating GitHub issue titles.", prompt, summary)
}

func (c *ClaudeCli) IssueBody(input string, summary string) (string, error) {
	prompt := fmt.Sprintf(IssueBody, input)
	return c.send("You are a helpful assistant writing GitHub issue descriptions.", prompt, summary)
}

func (c *ClaudeCli) IssueLabels(issue string, available []string) ([]string, error) {
	alllabels := strings.Join(available, ", ")
	prompt := fmt.Sprintf(Labels, issue, alllabels)
	resp, err := c.send("You are a helpful assistant assigning GitHub issue labels.", prompt, "")
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

func (c *ClaudeCli) CommitMessage(number, diff, descr string) (string, error) {
	prompt := appendIssue(fmt.Sprintf(CommitMsg, trimInput(diff), number, number), descr)
	c.log.Debug("claude cli prompt: %q", prompt)
	return c.send("You are a helpful assistant writing commit messages.", prompt, "")
}

func (c *ClaudeCli) Summary(readme string) (string, error) {
	prompt := fmt.Sprintf(Summary, trimInput(readme))
	return c.send("You are a helpful assistant writing project summaries.", prompt, "")
}

func (c *ClaudeCli) SuggestBranch(descr string) (string, error) {
	prompt := fmt.Sprintf(BranchName, descr)
	return c.send("You are a helpful assistant suggesting branch names.", prompt, "")
}

func (c *ClaudeCli) send(system, user, summary string) (string, error) {
	content := user
	if c.summary {
		content = appendSummary(content, summary)
	}
	content = appendLanguage(content, c.language)
	path, err := exec.LookPath(c.binary)
	if err != nil {
		return "", fmt.Errorf("'%s' not found in PATH, install Claude Code to use the 'claude-cli' provider: %w", c.binary, err)
	}
	cmd := exec.Command(path, "-p", "--no-session-persistence", "--system-prompt", system, "--tools", "", "--model", c.model, "--output-format", "text")
	cmd.Env = withoutAPIKey(os.Environ())
	cmd.Stdin = strings.NewReader(content)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running claude cli: %v; stderr: %s", err, strings.TrimSpace(stderr.String()))
	}
	out := strings.TrimSpace(stdout.String())
	if out == "" {
		return "", errors.New("no content in claude cli response")
	}
	return out, nil
}

func withoutAPIKey(env []string) []string {
	var res []string
	for _, kv := range env {
		if !strings.HasPrefix(kv, "ANTHROPIC_API_KEY=") {
			res = append(res, kv)
		}
	}
	return res
}
