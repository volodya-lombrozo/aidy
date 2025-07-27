package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name:     "With summary",
			prompt:   "This is a prompt.",
			summary:  "This is a summary.",
			expected: "This is a prompt.\nThis is the project summary for which you do it:\n<summary>\nThis is a summary.\n</summary>\n",
		},
		{
			name:     "Empty prompt with summary",
			prompt:   "",
			summary:  "This is a summary.",
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
			result := appendSummary(tt.prompt, tt.summary)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrimPrompt(t *testing.T) {
	prompt := largePrompt()

	trimmed := trimPrompt(prompt)

	assert.Equal(t, 48_000, len(trimmed))
}

func TestDoesNotTrimPrompt(t *testing.T) {
	prompt := "Simple Prompt"

	trimmed := trimPrompt(prompt)

	assert.Equal(t, "Simple Prompt", trimmed)
}

func largePrompt() string {
	return strings.Repeat(string('a'), 100*500)
}

func TestAppendIssue_WhenDescriptionIsEmpty_ReturnsPromptUnchanged(t *testing.T) {
	prompt := "Do something important"
	desc := ""
	result := appendIssue(prompt, desc)

	assert.Equal(t, prompt, result)
}

func TestAppendIssue_WhenDescriptionIsNotEmpty_AppendsFormattedDescription(t *testing.T) {
	prompt := "Do something important"
	desc := "This is a test issue description"
	expected := "Do something important\n" +
		"This is the issue description for which you do it:\n" +
		"<issue>\n" +
		"This is a test issue description\n" +
		"</issue>\n"
	result := appendIssue(prompt, desc)

	assert.Equal(t, expected, result)
}

func TestAppendIssue_WhenDescriptionContainsNewlines_AppendsDescriptionWithNewlinesPreserved(t *testing.T) {
	prompt := "Handle this"
	desc := "Line 1\nLine 2\nLine 3"
	expected := "Handle this\n" +
		"This is the issue description for which you do it:\n" +
		"<issue>\n" +
		"Line 1\nLine 2\nLine 3\n" +
		"</issue>\n"
	result := appendIssue(prompt, desc)

	assert.Equal(t, expected, result)
}

func TestAppendIssue_WithSpecialCharactersInDescription_AppendsCorrectly(t *testing.T) {
	prompt := "Process this"
	desc := "Issue with special chars: !@#$%^&*()"
	expected := "Process this\n" +
		"This is the issue description for which you do it:\n" +
		"<issue>\n" +
		"Issue with special chars: !@#$%^&*()\n" +
		"</issue>\n"
	result := appendIssue(prompt, desc)

	assert.Equal(t, expected, result)
}

func TestAppendIssue_WithEmptyPromptAndEmptyDescription_ReturnsEmptyString(t *testing.T) {
	prompt := ""
	desc := ""
	result := appendIssue(prompt, desc)

	assert.Equal(t, "", result)
}

func TestAppendIssue_WithEmptyPromptAndNonEmptyDescription_AppendsFormattedDescription(t *testing.T) {
	prompt := ""
	desc := "Non-empty description"
	expected := "\n" +
		"This is the issue description for which you do it:\n" +
		"<issue>\n" +
		"Non-empty description\n" +
		"</issue>\n"
	result := appendIssue(prompt, desc)

	assert.Equal(t, expected, result)
}

func TestAppendIssue_WhenPromptContainsSpecialCharacters_CombinesWithDescriptionProperly(t *testing.T) {
	prompt := "Prompt with special chars: ~`|\\<>"
	desc := "Description test"
	expected := "Prompt with special chars: ~`|\\<>\n" +
		"This is the issue description for which you do it:\n" +
		"<issue>\n" +
		"Description test\n" +
		"</issue>\n"
	result := appendIssue(prompt, desc)

	require.Equal(t, expected, result)
}
