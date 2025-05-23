package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestAppend_Help(t *testing.T) {
	var out bytes.Buffer
	command := newAppendCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "command should execute without error")
	assert.Contains(t, out.String(), "Append all local changes to the last commit")
}

func TestAppend_Execution(t *testing.T) {
	maidy := aidy.NewMock()
	ctx := &Context{Assistant: maidy}
	command := newAppendCmd(ctx)

	err := command.Execute()

	require.NoError(t, err, "command should execute without error")
	assert.Contains(t, maidy.Logs(), "Append called")
}
