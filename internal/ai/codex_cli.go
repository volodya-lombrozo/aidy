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

const codexCliDefaultModel = "gpt-5-codex"
const codexCliBinary = "codex"

type CodexCli struct {
	binary   string
	model    string
	summary  bool
	language string
	log      log.Logger
}

func NewCodexCli(model string, summary bool, language string) AI {
	if model == "" {
		model = codexCliDefaultModel
	}
	return &CodexCli{
		binary:   codexCliBinary,
		model:    model,
		summary:  summary,
		language: language,
		log:      log.Default(),
	}
}

func (c *CodexCli) ReleaseNotes(changes string) (string, error) {
	return c.send("You are a helpful assistant generating GitHub release notes.", fmt.Sprintf(ReleaseNotes, trimInput(changes)), "")
}

func (c *CodexCli) PrTitle(number, diff, issue, summary string) (string, error) {
	return c.send("You are a helpful assistant generating Git commit titles.", fmt.Sprintf(PrTitle, trimInput(diff), issue, number, number), summary)
}

func (c *CodexCli) PrBody(diff string, issue string, summary string) (string, error) {
	return c.send("You are a helpful assistant generating Git commit messages.", fmt.Sprintf(PrBody, trimInput(diff), issue), summary)
}

func (c *CodexCli) IssueTitle(userInput string, summary string) (string, error) {
	return c.send("You are a helpful assistant creating GitHub issue titles.", fmt.Sprintf(IssueTitle, userInput), summary)
}

func (c *CodexCli) IssueBody(input string, summary string) (string, error) {
	return c.send("You are a helpful assistant writing GitHub issue descriptions.", fmt.Sprintf(IssueBody, input), summary)
}

func (c *CodexCli) IssueLabels(issue string, available []string) ([]string, error) {
	labels := strings.Join(available, ", ")
	response, err := c.send("You are a helpful assistant assigning GitHub issue labels.", fmt.Sprintf(Labels, issue, labels), "")
	if err != nil {
		return nil, err
	}
	var result []string
	for _, label := range available {
		if strings.Contains(response, label) {
			result = append(result, label)
		}
	}
	return result, nil
}

func (c *CodexCli) CommitMessage(number, diff, descr string) (string, error) {
	prompt := appendIssue(fmt.Sprintf(CommitMsg, trimInput(diff), number, number), descr)
	c.log.Debug("codex cli prompt: %q", prompt)
	return c.send("You are a helpful assistant writing commit messages.", prompt, "")
}

func (c *CodexCli) Summary(readme string) (string, error) {
	return c.send("You are a helpful assistant writing project summaries.", fmt.Sprintf(Summary, trimInput(readme)), "")
}

func (c *CodexCli) SuggestBranch(descr string) (string, error) {
	return c.send("You are a helpful assistant suggesting branch names.", fmt.Sprintf(BranchName, descr), "")
}

func (c *CodexCli) send(system, user, summary string) (string, error) {
	content := user
	if c.summary {
		content = appendSummary(content, summary)
	}
	content = appendLanguage(content, c.language)
	path, err := exec.LookPath(c.binary)
	if err != nil {
		return "", fmt.Errorf("'%s' not found in PATH, install Codex CLI to use the 'codex-cli' provider: %w", c.binary, err)
	}
	prompt := fmt.Sprintf("%s\n\n%s", system, content)
	cmd := exec.Command(path, "exec", "--ephemeral", "--sandbox", "read-only", "--model", c.model, "--config", "model_reasoning_effort=low", "-")
	cmd.Env = withoutOpenAIAPIKey(os.Environ())
	cmd.Stdin = strings.NewReader(prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running codex cli: %v; stderr: %s", err, strings.TrimSpace(stderr.String()))
	}
	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("no content in codex cli response")
	}
	return result, nil
}

func withoutOpenAIAPIKey(env []string) []string {
	var result []string
	for _, variable := range env {
		if !strings.HasPrefix(variable, "OPENAI_API_KEY=") {
			result = append(result, variable)
		}
	}
	return result
}
