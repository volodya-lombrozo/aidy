package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestIssue_Help(t *testing.T) {
	var out bytes.Buffer
	command := newIssueCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "Generate a GitHub issue using an AI-generated title, body, and labels")
}

func TestIssue_Execution(t *testing.T) {
	mock := aidy.NewMock()
	ctx := &Context{Assistant: mock}
	command := newIssueCmd(ctx)
	command.SetArgs([]string{"test-description"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, mock.Logs(), "Issue called with task: test-description")
}
