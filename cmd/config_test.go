package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestConfig_Help(t *testing.T) {
	var out bytes.Buffer
	command := newConfigCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "Print the current configuration")
}

func TestConfig_Executes(t *testing.T) {
	var out bytes.Buffer
	mock := aidy.NewMock()
	command := newConfigCmd(&Context{Assistant: mock})
	command.SetOut(&out)

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, mock.Logs(), "PrintConfig called")
}
