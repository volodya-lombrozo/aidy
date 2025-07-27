package output

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/executor"
)

func TestEditor_Print_RunOption(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	editor.in = r
	_, err := io.WriteString(w, "r\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = editor.Print(command)

	require.NoError(t, err, "Print should not return an error")
	assert.Len(t, shell.Commands, 1, "expected 1 command to be run")
	assert.Equal(t, "echo 'Hello, World!'", shell.Commands[0], "expected command to match")
}

func TestEditor_Print_RunOption_Error(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("simulated error")
	editor := NewEditor(shell)
	editor.in = r
	_, err := io.WriteString(w, "r\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = editor.Print(command)

	assert.Error(t, err, "expected an error when running the command")
	assert.Equal(t, "simulated error", err.Error(), "expected error message to match")
}

func TestEditor_Print_PrintOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	editor.in = input_r
	editor.out = output_w
	_, err := io.WriteString(input_w, "p\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = editor.Print(command)

	assert.NoError(t, err, "Print should not return an error")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), command, "expected command in output")
	assert.Len(t, shell.Commands, 0, "expected no command to be run")
}

func TestEditor_Print_CancelOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	editor.in = input_r
	editor.out = output_w
	_, err := io.WriteString(input_w, "c\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	_ = editor.Print(command)

	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), "canceled", "expected cancel message in output")
	assert.Len(t, shell.Commands, 0, "expected no command to be run")
}

func TestEditor_Print_EditOption(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	editor.in = r
	_, err := io.WriteString(w, "e\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	_ = editor.Print(command)

	assert.Len(t, shell.Commands, 2, "expected 2 commands to be run")
	assert.Equal(t, command, shell.Commands[1], "expected edited command to match")
}

func TestEditor_Print_EditOption_FailsWithError(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	editor := NewEditor(shell)
	shell.Err = fmt.Errorf("simulated error")
	editor.in = r
	_, err := io.WriteString(w, "e\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = editor.Print(command)

	assert.Error(t, err, "expected an error when running the command")
	assert.Contains(t, err.Error(), "simulated error", "expected error message to match")
	assert.Contains(t, err.Error(), "failed to run command", "expected error to mention command failure")

}

func TestEditor_PrettyCommand(t *testing.T) {
	tests := []struct{ input, want string }{
		{"git commit --amend --no-edit", "git commit\n  --amend\n  --no-edit"},
		{"docker run --rm -it ubuntu bash", "docker run\n  --rm -it ubuntu bash"},
		{"--version", "\n  --version"},
		{"echo hello world", "echo hello world"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := prettyCommand(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEditor_FindEditor(t *testing.T) {
	original := os.Getenv("EDITOR")
	defer func() {
		if err := os.Setenv("EDITOR", original); err != nil {
			t.Errorf("failed to reset EDITOR environment variable: %v", err)
		}
	}()
	tests := []struct {
		envEditor string
		os        string
		expected  string
	}{
		{"", "windows", "notepad"},
		{"", "darwin", "vi"},
		{"", "linux", "vi"},
		{"nano", "linux", "nano"},
		{"code", "darwin", "code"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("EDITOR=%s OS=%s", test.envEditor, test.os), func(t *testing.T) {
			if err := os.Setenv("EDITOR", test.envEditor); err != nil {
				require.NoError(t, err, "failed to set EDITOR environment variable")
			}
			editor := findEditor(test.os)
			assert.Equal(t, test.expected, editor)
		})
	}
}

func TestEditor_CleanQoutes(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{`"hello"`, `"world"`}, []string{"hello", "world"}},
		{[]string{`"foo"`, `"bar"`, `"baz"`}, []string{"foo", "bar", "baz"}},
		{[]string{`"quoted"`, `unquoted`, `"another"`}, []string{"quoted", "unquoted", "another"}},
		{[]string{`""`, `"empty"`}, []string{"", "empty"}},
		{[]string{`"noquotes"`}, []string{"noquotes"}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.input), func(t *testing.T) {
			result := cleanQoutes(test.input)
			require.Equal(t, test.expected, result, "For input %v, expected %v, got %v", test.input, test.expected, result)
		})
	}
}

func TestEditor_SplitCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		// Test case 1: Simple command
		{"echo hello", []string{"echo", "hello"}},
		{`echo "hello world"`, []string{"echo", "\"hello world\""}},
		{`echo \"hello\"`, []string{"echo", `"hello"`}},
		{`echo "hello 'world'"`, []string{"echo", "\"hello 'world'\""}},
		{`echo hello\ world`, []string{"echo", "hello world"}},
		{"echo    hello", []string{"echo", "hello"}},
		{`echo "hello`, nil}, // This should panic
		{`echo hello\\world`, []string{"echo", "hello\\world"}},
		{`echo ""`, []string{"echo", "\"\""}},
		{`echo $PATH`, []string{"echo", "$PATH"}},
		{"git commit\n  --amend\n  --no-edit", []string{"git", "commit", "--amend", "--no-edit"}},
		{"git commit\n  --message \"hello\nworld\"", []string{"git", "commit", "--message", "\"hello\nworld\""}},
		{"gh issue create --title \"tests for \\`edit\\` and \\`run\\`\" --body \"new \\`edit\\` and \\`run\\` handling.\" --label \"enh,good first issue\" --repo v/a",
			[]string{"gh", "issue", "create", "--title", "\"tests for `edit` and `run`\"", "--body", "\"new `edit` and `run` handling.\"", "--label", "\"enh,good first issue\"", "--repo", "v/a"}},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					require.Nil(t, test.expected, "Unexpected panic for input %q", test.input)
				}
			}()
			result := splitCommand(test.input)
			if test.expected == nil {
				t.Errorf("Expected panic for input %q", test.input)
			} else {
				require.Equal(t, test.expected, result, "For input %q, expected %v, got %v", test.input, test.expected, result)
			}
		})
	}
}
