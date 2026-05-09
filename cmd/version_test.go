package cmd

import (
	"bytes"
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
