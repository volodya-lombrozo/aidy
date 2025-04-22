package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockGetBaseBranchName(t *testing.T) {
	git := MockGit{}
	output, err := git.GetBaseBranchName()
	if err != nil {
		panic(err)
	}
	if output != "main" {
		t.Fatal("Expected the branch name 'main'")
	}
}

func TestMockGetCurrentCommitMessage(t *testing.T) {
	git := MockGit{}
	output, err := git.GetCurrentCommitMessage()
	if err != nil {
		panic(err)
	}
	expected := "feat(#42): current commit message"
	if output != expected {
		t.Fatalf("Expected the commit message '%v', but got '%v'", expected, output)
	}
}

func TestMockGetBranchName(t *testing.T) {
	git := MockGit{}
	output, err := git.GetBranchName()
	if err != nil {
		panic(err)
	}
	if output != "41_working_branch" {
		t.Fatal("Expected the branch name 'main'")
	}
}

func TestMockGetDiff(t *testing.T) {
	git := MockGit{}
	output, err := git.GetDiff()
	if err != nil {
		panic(err)
	}
	if output != "mock-diff" {
		t.Fatal("Expected the diff 'mock-diff'")
	}
}

func TestMockGetCurrentDiff(t *testing.T) {
	git := MockGit{}
	output, err := git.GetCurrentDiff()
	if err != nil {
		panic(err)
	}
	if output != "current-mock-diff" {
		t.Fatal("Expected the diff 'mock-diff'")
	}
}

func TestMockGetAllRemoteURLs(t *testing.T) {
	git := MockGit{}

	output, err := git.Remotes()

	require.NoError(t, err)
	first := "https://github.com/volodya-lombrozo/aidy.git"
	second := "https://github.com/volodya-lombrozo/forked-aidy.git"
	assert.Equal(t, first, output[0])
	assert.Equal(t, second, output[1])
}

func TestMockGitInstalled(t *testing.T) {
    git := MockGit{}

    installed, err  := git.Installed()

    require.NoError(t,err)
    assert.True(t, installed)
}
