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

func TestMock_Captured(t *testing.T) {
	mock := NewMock()
	require.Equal(t, "", mock.Captured(), "expected empty string when nothing was captured")

	_ = mock.Print("command1")
	_ = mock.Print("command2")

	require.Equal(t, "command1\ncommand2", mock.Captured(), "expected captured commands joined by newline")
}

func TestMock_Edit_DefaultAccepts(t *testing.T) {
	mock := NewMock()

	result, err := mock.Edit("some text")

	require.NoError(t, err, "expected no error by default")
	require.Equal(t, "some text", result, "expected the original text to be returned unchanged")
	require.Equal(t, "some text", mock.Captured(), "expected the text to be recorded")
}

func TestMock_Edit_ReturnsConfiguredText(t *testing.T) {
	mock := NewMock()
	mock.EditText = "edited text"

	result, err := mock.Edit("some text")

	require.NoError(t, err, "expected no error")
	require.Equal(t, "edited text", result, "expected the configured text to be returned")
}

func TestMock_Edit_ReturnsConfiguredError(t *testing.T) {
	mock := NewMock()
	mock.EditErr = ErrCanceled

	result, err := mock.Edit("some text")

	require.ErrorIs(t, err, ErrCanceled, "expected the configured error to be returned")
	require.Equal(t, "", result, "expected no text to be returned on error")
}
