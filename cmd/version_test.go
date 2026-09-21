package cmd

import (
	"bytes"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCmd_PrintsVersion(t *testing.T) {
	var out bytes.Buffer
	command := newVersionCmd()
	command.SetOut(&out)

	err := command.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), resolvedVersion())
}

func TestResolveVersion_PrefersLdflagOverBuildInfo(t *testing.T) {
	got := resolveVersion("v0.2.1", func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, true
	})
	assert.Equal(t, "v0.2.1", got)
}

func TestResolveVersion_UsesGoInstallModuleVersion(t *testing.T) {
	got := resolveVersion("dev", func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.2.1"}}, true
	})
	assert.Equal(t, "v0.2.1", got)
}

func TestResolveVersion_KeepsDevForLocalDevelBuilds(t *testing.T) {
	got := resolveVersion("dev", func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
	})
	assert.Equal(t, "dev", got)
}

func TestResolveVersion_KeepsDevWhenBuildInfoMissing(t *testing.T) {
	got := resolveVersion("dev", func() (*debug.BuildInfo, bool) {
		return nil, false
	})
	assert.Equal(t, "dev", got)
}

func TestResolveVersion_EmptyLdflagFallsBackToModule(t *testing.T) {
	got := resolveVersion("", func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.2.1"}}, true
	})
	assert.Equal(t, "v0.2.1", got)
}

func TestVersionCmd_Help(t *testing.T) {
	var out bytes.Buffer
	command := newVersionCmd()
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Print the current aidy version")
}
