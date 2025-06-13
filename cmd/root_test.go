package cmd

import (
	"bytes"
	"testing"

	"github.com/volodya-lombrozo/aidy/internal/aidy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_PrintsHelp(t *testing.T) {
	var out bytes.Buffer
	command := newRootCmd(mock)
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "Aidy assists you with generating commit messages, pull requests, issues, and releases")
}

func TestRootCmd_Executes_WithoutError(t *testing.T) {
	err := Execute()

	assert.NoError(t, err, "no error expected")
}

func mock(summary, aider, ailess, silent, debug bool) aidy.Aidy {
	return aidy.NewMock()
}
