package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func TestStart_Help(t *testing.T) {
	var out bytes.Buffer
	command := newStartCmd(&Context{})
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "Start a new issue")
}

func TestStart_Execution(t *testing.T) {
	mock := aidy.NewMock()
	ctx := &Context{Assistant: mock}
	command := newStartCmd(ctx)
	command.SetArgs([]string{"123"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, mock.Logs(), "StartIssue called with number: 123")
}
