package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrettyCommand(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "git commit --amend --no-edit",
			want:  "git commit\n  --amend\n  --no-edit",
		},
		{
			input: "docker run --rm -it ubuntu bash",
			want:  "docker run\n  --rm -it ubuntu bash",
		},
		{
			input: "--version",
			want:  "\n  --version",
		},
		{
			input: "echo hello world",
			want:  "echo hello world",
		},
		{
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := prettyCommand(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindEditor(t *testing.T) {
	originalEditor := os.Getenv("EDITOR")
	defer func() {
		if err := os.Setenv("EDITOR", originalEditor); err != nil {
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
				t.Errorf("failed to set EDITOR environment variable: %v", err)
			}
			editor := findEditor(test.os)
			assert.Equal(t, test.expected, editor)
		})
	}
}

func TestMockOutput(t *testing.T) {
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

func TestCleanQoutes(t *testing.T) {
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

func TestPrinter_Print(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printer := NewPrinter()

	_ = printer.Print("hello")
    err := w.Close();
    require.NoError(t, err, "failed to close write pipe")
	var buf bytes.Buffer
    _, err = io.Copy(&buf, r);
    require.NoError(t, err, "failed to copy data")
	os.Stdout = originalStdout
	expected := "hello\n"
	require.Equal(t, expected, buf.String(), "Output should match expected value")
}

func TestSplitCommand(t *testing.T) {
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
