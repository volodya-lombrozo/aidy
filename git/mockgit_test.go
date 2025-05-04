package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/executor"
)

func TestMockGetBaseBranchName(t *testing.T) {
	git := MockGit{}
	output, err := git.GetBaseBranchName()

	require.NoError(t, err)
	assert.Equal(t, "main", output)
}

func TestMockGitRoot(t *testing.T) {
	git := MockGit{}

	root, err := git.Root()
	require.NoError(t, err)
	assert.Equal(t, "/dev/null", root)
}

func TestMockGitWithDirReset(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	err := git.Reset("HEAD~1")
	require.NoError(t, err)
}

func TestMockGitWithDirGetBaseBranchName(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	output, err := git.GetBaseBranchName()
	require.NoError(t, err)
	assert.Equal(t, "main", output)
}

func TestMockGitWithDirAppendToCommit(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	err := git.AppendToCommit()
	require.NoError(t, err)
}

func TestMockGitWithDirGetBranchName(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	output, err := git.GetBranchName()
	require.NoError(t, err)
	assert.Equal(t, "41_working_branch", output)
}

func TestMockGitWithDirGetDiff(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	output, err := git.GetDiff()
	require.NoError(t, err)
	assert.Equal(t, "mock-diff", output)
}

func TestMockGitWithDirGetCurrentDiff(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	output, err := git.GetCurrentDiff()
	require.NoError(t, err)
	assert.Equal(t, "current-mock-diff", output)
}

func TestMockGitWithDirCommitChanges(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	err := git.CommitChanges("Initial commit")
	require.NoError(t, err)
}

func TestMockGitWithDirGetCurrentCommitMessage(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	output, err := git.GetCurrentCommitMessage()
	require.NoError(t, err)
	assert.Equal(t, "feat(#42): current commit message", output)
}

func TestMockGitWithDirRemotes(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	output, err := git.Remotes()
	require.NoError(t, err)
	assert.Equal(t, []string{"https://github.com/volodya-lombrozo/aidy.git", "https://github.com/volodya-lombrozo/forked-aidy.git"}, output)
}

func TestMockGitWithDirInstalled(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	installed, err := git.Installed()
	require.NoError(t, err)
	assert.True(t, installed)
}

func TestMockGitWithDirRoot(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	root, err := git.Root()
	require.NoError(t, err)
	assert.Equal(t, "/some/dir", root)
}

func TestMockGitWithDirAddAll(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	err := git.AddAll()
	require.NoError(t, err)
}

func TestMockGitWithDirAmend(t *testing.T) {
	git := NewMockGitWithDir("/some/dir")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	err := git.Amend("new message")
	require.NoError(t, err)
}

func TestMockGitAmendSuccess(t *testing.T) {
	executor := &executor.MockExecutor{}
	git := MockGit{Shell: executor}

	err := git.Amend("new message")
	require.NoError(t, err)
	assert.Contains(t, executor.Commands, "git commit --amend -m new message ")
}

func TestMockGitAmendError(t *testing.T) {
	executor := &executor.MockExecutor{Err: fmt.Errorf("amend error")}
	git := MockGit{Shell: executor}

	err := git.Amend("new message")
	require.Error(t, err)
	assert.Equal(t, "amend error", err.Error())
}

func TestMockGitAddAllSuccess(t *testing.T) {
	executor := &executor.MockExecutor{}
	git := MockGit{Shell: executor}

	err := git.AddAll()
	require.NoError(t, err)
	assert.Contains(t, executor.Commands, "git add --all ")
}

func TestMockGitAddAllError(t *testing.T) {
	executor := &executor.MockExecutor{Err: fmt.Errorf("add all error")}
	git := MockGit{Shell: executor}

	err := git.AddAll()
	require.Error(t, err)
	assert.Equal(t, "add all error", err.Error())
}

func TestMockGitResetSuccess(t *testing.T) {
	executor := &executor.MockExecutor{}
	git := MockGit{Shell: executor}

	err := git.Reset("HEAD~1")
	require.NoError(t, err)
	assert.Contains(t, executor.Commands, "git reset --soft HEAD~1")
}

func TestMockGitResetError(t *testing.T) {
	executor := &executor.MockExecutor{Err: fmt.Errorf("reset error")}
	git := MockGit{Shell: executor}

	err := git.Reset("HEAD~1")
	require.Error(t, err)
	assert.Equal(t, "reset error", err.Error())
}

func TestMockGitAppendToCommit(t *testing.T) {
	git := MockGit{}

	err := git.AppendToCommit()
	require.NoError(t, err)
}

func TestMockGitCommitChanges(t *testing.T) {
	git := MockGit{}

	err := git.CommitChanges("Initial commit")
	require.NoError(t, err)
}

func TestMockGetCurrentCommitMessage(t *testing.T) {
	git := MockGit{}
	output, err := git.GetCurrentCommitMessage()

	require.NoError(t, err)
	expected := "feat(#42): current commit message"
	assert.Equal(t, expected, output)
}

func TestMockGetBranchName(t *testing.T) {
	git := MockGit{}
	output, err := git.GetBranchName()

	require.NoError(t, err)
	assert.Equal(t, "41_working_branch", output)
}

func TestMockGetDiff(t *testing.T) {
	git := MockGit{}
	output, err := git.GetDiff()

	require.NoError(t, err)
	assert.Equal(t, "mock-diff", output)
}

func TestMockGetCurrentDiff(t *testing.T) {
	git := MockGit{}
	output, err := git.GetCurrentDiff()

	require.NoError(t, err)
	assert.Equal(t, "current-mock-diff", output)
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

	installed, err := git.Installed()

	require.NoError(t, err)
	assert.True(t, installed)
}
