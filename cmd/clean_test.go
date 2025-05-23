package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestClean_Help(t *testing.T) {
	var out bytes.Buffer
	command := newCleanCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "command should execute without error")
	assert.Contains(t, out.String(), "Clean the aidy cache")
}

func TestClean_Execution(t *testing.T) {
	maidy := aidy.NewMock()
	ctx := &Context{Assistant: maidy}
	command := newCleanCmd(ctx)

	err := command.Execute()

	require.NoError(t, err, "command should execute without error")
	assert.Contains(t, maidy.Logs(), "Clean called")
}
