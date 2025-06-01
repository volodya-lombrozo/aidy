package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockGithub_IssueDescription(t *testing.T) {
	mock := NewMock()
	number := "123"
	expected := "Mock description for issue #" + number

	description, err := mock.Description(number)

	require.NoError(t, err, "mock object should not return errors")
	assert.Equal(t, expected, description, "Description should match expected value")
}

func TestMockGithub_Labels(t *testing.T) {
	mock := NewMock()

	labels, err := mock.Labels()

	require.NoError(t, err, "mock object should not return errors")
	assert.Contains(t, labels, "bug", "Labels should contain 'bug'")
	assert.Contains(t, labels, "documentation", "Labels should contain 'documentation'")
	assert.Contains(t, labels, "question", "Labels should contain 'question'")
}

func TestMockGithub_Remotes(t *testing.T) {
	expected := []string{"volodya-lombrozo/aidy", "volodya-lombrozo/jtcop"}
	gh := NewMock()

	acutal, err := gh.Remotes()

	require.NoError(t, err, "mock object should not return errors")
	assert.Equal(t, expected, acutal)
}
