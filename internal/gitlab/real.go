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
	out, err := r.shell.RunCommand("glab", "mr", "list", "--source-branch", branch, "--output", "json")
	if err != nil {
		return "", "", fmt.Errorf("error fetching merge requests for branch '%s': %w", branch, err)
	}
	var mrs []mergeRequest
	if err := json.Unmarshal([]byte(out), &mrs); err != nil {
		return "", "", fmt.Errorf("error parsing merge request json for branch '%s': %w", branch, err)
	}
	if len(mrs) == 0 {
		return "", "", fmt.Errorf("no open merge request found for branch '%s'", branch)
	}
	if len(mrs) > 1 {
		r.log.Warn("found %d open merge requests for branch '%s', using the first one: '%s'", len(mrs), branch, mrs[0].Title)
	}
	mr := mrs[0]
	r.log.Debug("found an open merge request '%s' for branch '%s'", mr.Title, branch)
	return mr.Title, mr.Description, nil
}
