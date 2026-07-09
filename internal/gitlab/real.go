package gitlab

import (
	"encoding/json"
	"fmt"

	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/log"
)

type real struct {
	shell executor.Executor
	log   log.Logger
}

type mergeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewGitlab(shell executor.Executor) *real {
	return &real{shell: shell, log: log.Default()}
}

func (r *real) MergeRequestByBranch(branch string) (string, string, error) {
	out, err := r.shell.RunCommand("glab", "mr", "view", branch, "--output", "json")
	if err != nil {
		return "", "", fmt.Errorf("error fetching merge request for branch '%s': %w", branch, err)
	}
	var mr mergeRequest
	if err := json.Unmarshal([]byte(out), &mr); err != nil {
		return "", "", fmt.Errorf("error parsing merge request json for branch '%s': %w", branch, err)
	}
	if mr.Title == "" {
		return "", "", fmt.Errorf("no open merge request found for branch '%s'", branch)
	}
	r.log.Debug("found an open merge request '%s' for branch '%s'", mr.Title, branch)
	return mr.Title, mr.Description, nil
}
