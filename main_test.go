package main

import (
    "bytes"
    "io"
    "os"
    "strings"
    "testing"
    "github.com/volodya-lombrozo/aidy/ai"
)


func TestHandleIssue(t *testing.T) {
    mockAI := &ai.MockAI{}
    userInput := "test input"
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    handleIssue(userInput, mockAI)
    w.Close()
    os.Stdout = old
    var buf bytes.Buffer
    io.Copy(&buf, r)
    output := buf.String()
    expected := "Generated Issue Command:\ngh issue create --title \"Mock Issue Title for test input\" --body \"Mock Issue Body for test input\""
    if strings.TrimSpace(output) != strings.TrimSpace(expected) {
        t.Errorf("Unexpected output:\n%s", output)
    }
}


func TestHandleHelp(t *testing.T) {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    handleHelp()
    w.Close()
    os.Stdout = old
    var buf bytes.Buffer
    io.Copy(&buf, r)
    output := buf.String()
    expected := `Usage:
  aidy pr   - Generate a pull request using AI-generated title and body.
  aidy help - Show this help message.`
    if strings.TrimSpace(output) != strings.TrimSpace(expected) {
        t.Errorf("Unexpected output:\n%s", output)
    }
}

func TestExtractIssueNumber(t *testing.T) {
    tests := []struct {
        branchName string
        expected   string
    }{
        {"123_feature", "123"},
        {"456_bugfix", "456"},
        {"789", "789"},
        {"no_issue_number", "no"},
        {"", "unknown"},
    }

    for _, test := range tests {
        result := extractIssueNumber(test.branchName)
        if result != test.expected {
            t.Errorf("For branch name '%s', expected '%s', got '%s'", test.branchName, test.expected, result)
        }
    }
}

func TestEscapeBackticks(t *testing.T) {
    input := "This is a `test` string with `backticks`."
    expected := "This is a \\`test\\` string with \\`backticks\\`."
    result := escapeBackticks(input)

    if result != expected {
        t.Fatalf("Expected '%s', got '%s'", expected, result)
    }
}
