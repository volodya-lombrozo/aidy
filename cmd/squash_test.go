package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestSquash_Help(t *testing.T) {
	var out bytes.Buffer
	command := newSquashCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "Squash all commits in the current branch into a single commit")
}

func TestSquash_Execution(t *testing.T) {
	mock := aidy.NewMock()
	ctx := &Context{Assistant: mock}
	command := newSquashCmd(ctx)

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, mock.Logs(), "Squash called")
}
