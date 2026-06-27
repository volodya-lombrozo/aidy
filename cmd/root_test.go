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
	command := NewRootCmd(mock)
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

func TestRootCmd_PrintsUsageOnBadInput(t *testing.T) {
	var out bytes.Buffer
	command := NewRootCmd(mock)
	command.SetOut(&out)
	command.SetArgs([]string{"commit", "--unknown-flag"})

	_ = command.Execute()

	assert.Contains(t, out.String(), "Usage:", "usage should be printed on bad user input")
}

func TestRootCmd_SilencesUsageOnRuntimeError(t *testing.T) {
	failing := func(summary, aider, ailess, silent, debug bool, language string) aidy.Aidy {
		return aidy.NewFailingMock()
	}
	var out bytes.Buffer
	command := NewRootCmd(failing)
	command.SetOut(&out)
	command.SetArgs([]string{"commit"})

	_ = command.Execute()

	assert.NotContains(t, out.String(), "Usage:", "usage should not be printed on runtime errors")
}

func mock(summary, aider, ailess, silent, debug bool, language string) aidy.Aidy {
	return aidy.NewMock()
}
