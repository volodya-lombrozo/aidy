package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
