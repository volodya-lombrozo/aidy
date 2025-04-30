package output

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

func TestPrinter_Print(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printer := NewPrinter()

	printer.Print("hello")

	if err := w.Close(); err != nil {
		t.Errorf("failed to close write pipe: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Errorf("failed to copy data: %v", err)
	}
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
		{`echo "hello world"`, []string{"echo", "hello world"}},
		{`echo \"hello\"`, []string{"echo", `"hello"`}},
		{`echo "hello 'world'"`, []string{"echo", "hello 'world'"}},
		{`echo hello\ world`, []string{"echo", "hello world"}},
		{"echo    hello", []string{"echo", "hello"}},
		{`echo "hello`, nil}, // This should panic
		{`echo hello\\world`, []string{"echo", "hello\\world"}},
		{`echo ""`, []string{"echo"}},
		{`echo $PATH`, []string{"echo", "$PATH"}},
		{"gh issue create --title \"tests for \\`edit\\` and \\`run\\`\" --body \"new \\`edit\\` and \\`run\\` handling.\" --label \"enh,good first issue\" --repo v/a",
			[]string{"gh", "issue", "create", "--title", "tests for `edit` and `run`", "--body", "new `edit` and `run` handling.", "--label", "enh,good first issue", "--repo", "v/a"}},
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
