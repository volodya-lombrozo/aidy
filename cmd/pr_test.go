package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestPr_Help(t *testing.T) {
	var out bytes.Buffer
	command := newPrCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "Create a PR based on changes in the current branch")
}

func TestPr_Execution(t *testing.T) {
	mock := aidy.NewMock()
	ctx := &Context{Assistant: mock}
	command := newPrCmd(ctx)

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, mock.Logs(), "PullRequest called")
}
