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

func TestRootCmd_LanguageShortFlagDoesNotSwallowDuplicate(t *testing.T) {
	command := NewRootCmd(mock)
	var errBuf bytes.Buffer
	command.SetErr(&errBuf)
	command.SetArgs([]string{"mr", "-f", "-l", "--duplicate"})

	err := command.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid language \"--duplicate\"")
	assert.Contains(t, err.Error(), "flag needs an argument")
}

func TestRootCmd_LanguageEqualsFlagValueIsRejected(t *testing.T) {
	command := NewRootCmd(mock)
	var errBuf bytes.Buffer
	command.SetErr(&errBuf)
	command.SetArgs([]string{"mr", "-f", "--language=--duplicate"})

	err := command.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid language \"--duplicate\"")
	assert.Contains(t, err.Error(), "flag needs an argument")
}

func TestRootCmd_LanguageThenDuplicateStillWorks(t *testing.T) {
	var gotLang string
	mockAidy := aidy.NewMock()
	factory := func(summary, aider, ailess, silent, debug bool, language string) aidy.Aidy {
		gotLang = language
		return mockAidy
	}
	command := NewRootCmd(factory)
	command.SetArgs([]string{"mr", "-f", "-l", "fr", "--duplicate"})

	err := command.Execute()

	require.NoError(t, err)
	assert.Equal(t, "fr", gotLang)
	assert.Contains(t, mockAidy.Logs(), "MergeRequest called")
	mr, _, findErr := command.Find([]string{"mr"})
	require.NoError(t, findErr)
	dup, dupErr := mr.Flags().GetBool("duplicate")
	require.NoError(t, dupErr)
	assert.True(t, dup)
}
