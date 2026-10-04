package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestLast_Help(t *testing.T) {
	var out bytes.Buffer
	command := newLastCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "command should execute without error")
	assert.Contains(t, out.String(), "Repeat the last generated command")
}

func TestLast_Execution(t *testing.T) {
	maidy := aidy.NewMock()
	command := newLastCmd(&Context{Assistant: maidy})

	err := command.Execute()

	require.NoError(t, err, "command should execute without error")
	assert.Contains(t, maidy.Logs(), "Last called")
}

func TestLast_Failure(t *testing.T) {
	command := newLastCmd(&Context{Assistant: aidy.NewFailingMock()})
	command.SilenceUsage = true
	command.SilenceErrors = true

	err := command.Execute()

	assert.Error(t, err, "command should report the failure")
}
