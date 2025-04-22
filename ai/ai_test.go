package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendSummary(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		summary  string
		expected string
	}{
		{
			name:     "No summary",
			prompt:   "This is a prompt.",
			summary:  "",
			expected: "This is a prompt.",
		},
		{
			name:   "With summary",
			prompt: "This is a prompt.",
			summary: "This is a summary.",
			expected: "This is a prompt.\nThis is the project summary for which you do it:\n<summary>\nThis is a summary.\n</summary>\n",
		},
		{
			name:   "Empty prompt with summary",
			prompt: "",
			summary: "This is a summary.",
			expected: "\nThis is the project summary for which you do it:\n<summary>\nThis is a summary.\n</summary>\n",
		},
		{
			name:     "Empty prompt and summary",
			prompt:   "",
			summary:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AppendSummary(tt.prompt, tt.summary)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrimPrompt(t *testing.T) {
	prompt := largePrompt()

	trimmed := TrimPrompt(prompt)

	assert.Equal(t, 48_000, len(trimmed))
}

func TestDoesNotTrimPrompt(t *testing.T) {
	prompt := "Simple Prompt"

	trimmed := TrimPrompt(prompt)

	assert.Equal(t, "Simple Prompt", trimmed)
}

func largePrompt() string {
	return strings.Repeat(string('a'), 100*500)
}
