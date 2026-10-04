package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSummarize_CountsRunsAndFailures(t *testing.T) {
	early := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	late := time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)
	runs := []Run{
		{Command: "commit", Time: late, Success: true},
		{Command: "commit", Time: early, Success: false},
		{Command: "start", Time: early, Success: true},
	}

	summaries := Summarize(runs, []string{"commit", "start", "squash"})

	assert.Equal(t, []Summary{
		{Command: "commit", Runs: 2, Failed: 1, Last: late},
		{Command: "start", Runs: 1, Failed: 0, Last: early},
		{Command: "squash"},
	}, summaries)
}

func TestSummarize_KeepsUnknownCommands(t *testing.T) {
	summaries := Summarize([]Run{{Command: "removed", Success: true}}, []string{"commit"})

	assert.Equal(t, []Summary{
		{Command: "removed", Runs: 1},
		{Command: "commit"},
	}, summaries)
}

func TestSummarize_SortsUnusedCommandsByName(t *testing.T) {
	summaries := Summarize(nil, []string{"start", "commit"})

	assert.Equal(t, []Summary{{Command: "commit"}, {Command: "start"}}, summaries)
}
