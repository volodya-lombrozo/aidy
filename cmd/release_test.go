package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestRelease_Help(t *testing.T) {
	var out bytes.Buffer
	command := newReleaseCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "Create a release based on a semver increment")
}

func TestRelease_Execution(t *testing.T) {
	mock := aidy.NewMock()
	ctx := &Context{Assistant: mock}
	command := newReleaseCmd(ctx)
	command.SetArgs([]string{"minor", "--repo", "test-repo"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, mock.Logs(), "Release called with interval: minor, repo: test-repo")
}
