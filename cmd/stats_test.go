package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/metrics"
)

func TestStatsCmd_Help(t *testing.T) {
	var out bytes.Buffer
	command := newStatsCmd(metrics.NewMock())
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Print how often each aidy command was used")
}

func TestStatsCmd_PrintsSummaryForAllCommands(t *testing.T) {
	used := time.Date(2026, 9, 29, 12, 0, 0, 0, time.Local)
	usage := metrics.NewMock(
		metrics.Run{Command: "commit", Time: used, Success: true},
		metrics.Run{Command: "commit", Time: used, Success: false},
	)
	var out bytes.Buffer
	command := NewRootCmd(mock, usage)
	command.SetOut(&out)
	command.SetArgs([]string{"stats"})

	err := command.Execute()

	require.NoError(t, err)
	assert.Regexp(t, `command\s+runs\s+failed\s+last used`, out.String())
	assert.Regexp(t, `commit\s+2\s+1\s+2026-09-29`, out.String())
	assert.Regexp(t, `squash\s+0\s+-\s+never`, out.String())
	assert.NotContains(t, out.String(), "completion")
	assert.NotContains(t, out.String(), "help")
}
