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
	assert.Contains(t, out.String(), "dev")
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

func TestResolveVersion_PrefersStampedVersion(t *testing.T) {
	version := resolveVersion("v1.2.3", buildInfoWith("v0.2.1"))

	assert.Equal(t, "v1.2.3", version)
}

func TestResolveVersion_UsesModuleVersionWhenNotStamped(t *testing.T) {
	version := resolveVersion(devVersion, buildInfoWith("v0.2.1"))

	assert.Equal(t, "v0.2.1", version)
}

func TestResolveVersion_UsesPseudoVersionForLocalBuild(t *testing.T) {
	pseudo := "v0.2.2-0.20261011074850-c27aaf1d2e3f+dirty"

	version := resolveVersion(devVersion, buildInfoWith(pseudo))

	assert.Equal(t, pseudo, version)
}

func TestResolveVersion_FallsBackToDevForDevelBuild(t *testing.T) {
	version := resolveVersion(devVersion, buildInfoWith("(devel)"))

	assert.Equal(t, devVersion, version)
}

func TestResolveVersion_FallsBackToDevWithoutBuildInfo(t *testing.T) {
	version := resolveVersion(devVersion, func() (*debug.BuildInfo, bool) { return nil, false })

	assert.Equal(t, devVersion, version)
}

func buildInfoWith(version string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: version}}, true
	}
}
