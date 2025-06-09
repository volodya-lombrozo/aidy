package output

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMock_Output(t *testing.T) {
	mock := NewMock()
	_ = mock.Print("command1")
	require.Equal(t, "command1", mock.Last(), "Last command should be 'command1'")

	_ = mock.Print("command2")
	require.Equal(t, "command2", mock.Last(), "Last command should be 'command2'")

	_ = mock.Print("command3")
	require.Equal(t, "command3", mock.Last(), "Last command should be 'command3'")
}

func TestMockOutput_Panics_WhenNoCommands(t *testing.T) {
	mock := NewMock()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when no commands are captured")
		}
	}()
	mock.Last()
}
