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

func TestReviewer_Print_RunOption(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewReviewer(shell, NewMockMemory())
	reviewer.in = r
	_, err := io.WriteString(w, "r\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = reviewer.Print(Fixed(command))

	require.NoError(t, err, "Print should not return an error")
	assert.Len(t, shell.Commands, 1, "expected 1 command to be run")
	assert.Equal(t, "echo 'Hello, World!'", shell.Commands[0], "expected command to match")
}

func TestReviewer_Print_RunOption_Error(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("simulated error")
	reviewer := NewReviewer(shell, NewMockMemory())
	reviewer.in = r
	_, err := io.WriteString(w, "r\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = reviewer.Print(Fixed(command))

	assert.Error(t, err, "expected an error when running the command")
	assert.Equal(t, "simulated error", err.Error(), "expected error message to match")
}

func TestReviewer_Print_PrintOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewReviewer(shell, NewMockMemory())
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "p\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = reviewer.Print(Fixed(command))

	assert.NoError(t, err, "Print should not return an error")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), command, "expected command in output")
	assert.Len(t, shell.Commands, 0, "expected no command to be run")
}

func TestReviewer_Print_CancelOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewReviewer(shell, NewMockMemory())
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "c\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	_ = reviewer.Print(Fixed(command))

	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(output), "canceled", "expected cancel message in output")
	assert.Len(t, shell.Commands, 0, "expected no command to be run")
}

func TestReviewer_Print_EditOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewReviewer(shell, NewMockMemory())
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "e\nr\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	_ = reviewer.Print(Fixed(command))

	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	output, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Len(t, shell.Commands, 2, "expected 2 commands to be run")
	assert.Equal(t, command, shell.Commands[1], "expected edited command to match")
	assert.Contains(t, string(output), "updated command", "expected updated command label in output")
	assert.Contains(t, string(output), command, "expected updated command content in output")
}

func TestReviewer_Print_EditOption_FailsWithError(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewReviewer(shell, NewMockMemory())
	shell.Err = fmt.Errorf("simulated error")
	reviewer.in = r
	_, err := io.WriteString(w, "e\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = reviewer.Print(Fixed(command))

	assert.Error(t, err, "expected an error when running the command")
	assert.Contains(t, err.Error(), "simulated error", "expected error message to match")
	assert.Contains(t, err.Error(), "failed to run command", "expected error to mention command failure")

}

func TestReviewer_Print_GenerateOption(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewReviewer(shell, NewMockMemory())
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "g\nr\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	generated := 0
	command := func() (string, error) {
		generated++
		return fmt.Sprintf("echo version-%d", generated), nil
	}

	err = reviewer.Print(command)

	require.NoError(t, err, "Print should not return an error")
	assert.Equal(t, 2, generated, "expected a second version to be generated on demand")
	require.Len(t, shell.Commands, 1, "expected 1 command to be run")
	assert.Equal(t, "echo version-2", shell.Commands[0], "expected the freshly generated command to be run")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	out, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(out), "[g]enerate", "expected the generate option to be offered")
	assert.Contains(t, string(out), "echo version-2", "expected the new command to be shown")
}

func TestReviewer_Print_FailsWhenTextFails(t *testing.T) {
	reviewer := NewReviewer(executor.NewMock(), NewMockMemory())
	command := func() (string, error) {
		return "", fmt.Errorf("simulated error")
	}

	err := reviewer.Print(command)

	require.Error(t, err, "expected an error when the generator fails")
	assert.Contains(t, err.Error(), "simulated error", "expected the generator error to be reported")
}

func TestReviewer_Print_FailsWhenRegenerationFails(t *testing.T) {
	r, w, _ := os.Pipe()
	reviewer := NewReviewer(executor.NewMock(), NewMockMemory())
	reviewer.in = r
	_, err := io.WriteString(w, "g\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	generated := 0
	command := func() (string, error) {
		generated++
		if generated > 1 {
			return "", fmt.Errorf("simulated error")
		}
		return "echo 'Hello, World!'", nil
	}

	err = reviewer.Print(command)

	require.Error(t, err, "expected an error when regeneration fails")
	assert.Contains(t, err.Error(), "simulated error", "expected the generator error to be reported")
}

func TestReviewer_Print_GenerateOption_KeepsFixedCommand(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	shell := executor.NewMock()
	reviewer := NewReviewer(shell, NewMockMemory())
	reviewer.in = input_r
	_, err := io.WriteString(input_w, "g\nr\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")
	command := "echo 'Hello, World!'"

	err = reviewer.Print(Fixed(command))

	require.NoError(t, err, "Print should not return an error")
	require.Len(t, shell.Commands, 1, "expected 1 command to be run")
	assert.Equal(t, command, shell.Commands[0], "expected a fixed command to stay the same when generated again")
}

func TestReviewer_Print_RemembersGeneratedCommand(t *testing.T) {
	r, w, _ := os.Pipe()
	memory := NewMockMemory()
	reviewer := NewReviewer(executor.NewMock(), memory)
	reviewer.in = r
	_, err := io.WriteString(w, "c\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")

	err = reviewer.Print(Fixed("gh pr create --title \"title\""))

	require.NoError(t, err, "Print should not return an error")
	assert.Equal(t, []string{"gh pr create\n  --title \"title\""}, memory.Commands, "expected the generated command to be remembered")
}

func TestReviewer_Print_RemembersCommandBeforeFailedRun(t *testing.T) {
	r, w, _ := os.Pipe()
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("network is unreachable")
	memory := NewMockMemory()
	reviewer := NewReviewer(shell, memory)
	reviewer.in = r
	_, err := io.WriteString(w, "r\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")

	err = reviewer.Print(Fixed("echo hello"))

	require.Error(t, err, "expected the failed run to be reported")
	assert.Equal(t, []string{"echo hello"}, memory.Commands, "expected the command to survive a failed run")
}

func TestReviewer_Print_RemembersEditedAndRegeneratedCommands(t *testing.T) {
	r, w, _ := os.Pipe()
	memory := NewMockMemory()
	reviewer := NewReviewer(executor.NewMock(), memory)
	reviewer.in = r
	_, err := io.WriteString(w, "e\ng\nc\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")
	generated := 0
	command := func() (string, error) {
		generated++
		return fmt.Sprintf("echo version-%d", generated), nil
	}

	err = reviewer.Print(command)

	require.NoError(t, err, "Print should not return an error")
	assert.Equal(t, []string{"echo version-1", "echo version-1", "echo version-2"}, memory.Commands, "expected every version of the command to be remembered")
}

func TestReviewer_Replay_OffersNoGeneration(t *testing.T) {
	input_r, input_w, _ := os.Pipe()
	output_r, output_w, _ := os.Pipe()
	shell := executor.NewMock()
	memory := NewMockMemory()
	reviewer := NewReviewer(shell, memory)
	reviewer.in = input_r
	reviewer.out = output_w
	_, err := io.WriteString(input_w, "g\nr\n")
	require.NoError(t, err, "failed to write to pipe")
	err = input_w.Close()
	require.NoError(t, err, "failed to close write pipe")

	err = reviewer.Replay("echo hello")

	require.NoError(t, err, "Replay should not return an error")
	require.Len(t, shell.Commands, 1, "expected the replayed command to be run")
	assert.Equal(t, "echo hello", shell.Commands[0], "expected the replayed command to be run unchanged")
	assert.Empty(t, memory.Commands, "expected an untouched replay not to be remembered again")
	err = output_w.Close()
	require.NoError(t, err, "failed to close output pipe")
	out, err := io.ReadAll(output_r)
	require.NoError(t, err, "failed to read from output")
	assert.Contains(t, string(out), "last command:", "expected the replayed command to be labeled")
	assert.NotContains(t, string(out), "[g]enerate", "expected no generate option for a replay")
}

func TestReviewer_Replay_RemembersEditedCommand(t *testing.T) {
	r, w, _ := os.Pipe()
	memory := NewMockMemory()
	reviewer := NewReviewer(executor.NewMock(), memory)
	reviewer.in = r
	_, err := io.WriteString(w, "e\nc\n")
	require.NoError(t, err, "failed to write to pipe")
	err = w.Close()
	require.NoError(t, err, "failed to close write pipe")

	err = reviewer.Replay("echo hello")

	require.NoError(t, err, "Replay should not return an error")
	assert.Equal(t, []string{"echo hello"}, memory.Commands, "expected the edited command to be remembered")
}

func TestReviewer_PrettyCommand(t *testing.T) {
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

func TestReviewer_CleanQoutes(t *testing.T) {
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

func TestReviewer_SplitCommand(t *testing.T) {
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
