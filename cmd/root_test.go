package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/volodya-lombrozo/aidy/internal/aidy"
	"github.com/volodya-lombrozo/aidy/internal/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_PrintsHelp(t *testing.T) {
	var out bytes.Buffer
	command := NewRootCmd(mock, metrics.NewMock())
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
	command := NewRootCmd(mock, metrics.NewMock())
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
	command := NewRootCmd(failing, metrics.NewMock())
	command.SetOut(&out)
	command.SetArgs([]string{"commit"})

	_ = command.Execute()

	assert.NotContains(t, out.String(), "Usage:", "usage should not be printed on runtime errors")
}

func mock(summary, aider, ailess, silent, debug bool, language string) aidy.Aidy {
	return aidy.NewMock()
}

func TestTracked_RecordsSuccessfulRun(t *testing.T) {
	usage := metrics.NewMock()
	command := NewRootCmd(mock, usage)
	command.SetArgs([]string{"ci"})

	err := Tracked(command, usage)

	require.NoError(t, err)
	runs, err := usage.Runs()
	require.NoError(t, err)
	require.Len(t, runs, 1, "one run should be recorded")
	assert.Equal(t, "commit", runs[0].Command, "aliases should be recorded by command name")
	assert.True(t, runs[0].Success)
}

func TestTracked_RecordsFailedRun(t *testing.T) {
	failing := func(summary, aider, ailess, silent, debug bool, language string) aidy.Aidy {
		return aidy.NewFailingMock()
	}
	usage := metrics.NewMock()
	command := NewRootCmd(failing, usage)
	command.SetArgs([]string{"commit"})

	err := Tracked(command, usage)

	require.Error(t, err)
	runs, err := usage.Runs()
	require.NoError(t, err)
	require.Len(t, runs, 1, "one run should be recorded")
	assert.False(t, runs[0].Success)
}

func TestTracked_SkipsRootHelpAndCompletion(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"help", "commit"}, {"completion", "bash"}, {"unknown"}} {
		usage := metrics.NewMock()
		command := NewRootCmd(mock, usage)
		command.SetOut(&bytes.Buffer{})
		command.SetArgs(args)

		_ = Tracked(command, usage)

		runs, err := usage.Runs()
		require.NoError(t, err)
		assert.Empty(t, runs, "no runs expected for %v", args)
	}
}

func TestUsage_DisabledByConfig(t *testing.T) {
	home := fakeHome(t)
	t.Chdir(home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".aidy.conf.yml"), []byte("metrics: false\n"), 0600))

	require.NoError(t, Usage().Record(metrics.Run{Command: "commit"}))

	assert.NoFileExists(t, filepath.Join(home, ".aidy", "metrics.json"), "metrics should not be recorded")
}

func TestUsage_EnabledWithoutConfig(t *testing.T) {
	home := fakeHome(t)
	t.Chdir(home)

	require.NoError(t, Usage().Record(metrics.Run{Command: "commit"}))

	assert.FileExists(t, filepath.Join(home, ".aidy", "metrics.json"), "metrics should be recorded")
}

func fakeHome(t *testing.T) string {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}
