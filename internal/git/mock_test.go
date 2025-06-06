package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/executor"
)

func TestMock_Tags_Absent(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "absent"
	git := NewMockWithShell(shell)

	res, err := git.Tags("does-not-matter")

	require.NoError(t, err)
	assert.Empty(t, res, "Expected no tags when output is 'absent'")
}

func TestMock_Tags_Error(t *testing.T) {
	git := NewMockWithError(fmt.Errorf("mock error"))

	_, err := git.Tags("does-not-matter")

	assert.Error(t, err)
	assert.Equal(t, "mock error", err.Error())
}

func TestMock_Log_Success(t *testing.T) {
	git := NewMock()

	output, err := git.Log("HEAD~1")

	require.NoError(t, err)
	assert.Equal(t, []string{
		"ci(#120): Update CI to use Ubuntu 24.04 and add .aidy to gitignore",
		"chore(deps): update dependency ruby to v3.4.3 (#117)",
	}, output)
}

func TestMock_Log_Error(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("log error")
	git := NewMockWithShell(shell)

	output, err := git.Log("HEAD~1")

	require.Error(t, err)
	assert.Equal(t, "log error", err.Error())
	assert.Nil(t, output)
}

func TestMock_Run_Successfully(t *testing.T) {
	shell := executor.NewMock()
	git := NewMockWithShell(shell)

	_, err := git.Run("status")

	require.NoError(t, err)
	assert.Equal(t, "git status", shell.Commands[0])
}

func TestMock_Run_Error(t *testing.T) {
	shell := executor.NewMock()
	shell.Err = fmt.Errorf("strange error")
	git := NewMockWithShell(shell)

	_, err := git.Run("status")

	require.Error(t, err)
	assert.Equal(t, "strange error", err.Error())
}

func TestMockWithDir_Root(t *testing.T) {
	tmp := "/dev/null"
	git := NewMockWithDir(tmp)

	root, err := git.Root()

	require.NoError(t, err)
	assert.Equal(t, tmp, root)
}

func TestMock_BaseBranch(t *testing.T) {
	git := NewMock()
	output, err := git.BaseBranch()

	require.NoError(t, err)
	assert.Equal(t, "main", output)
}

func TestMock_Root(t *testing.T) {
	git := NewMock()

	root, err := git.Root()
	require.NoError(t, err)
	assert.Equal(t, "/dev/null", root)
}

func TestMock_Amend_Success(t *testing.T) {
	executor := &executor.MockExecutor{}
	git := NewMockWithShell(executor)

	err := git.Amend("new message")
	require.NoError(t, err)
	assert.Contains(t, executor.Commands, "git commit --amend -m new message ")
}

func TestMock_Amend_Error(t *testing.T) {
	executor := &executor.MockExecutor{Err: fmt.Errorf("amend error")}
	git := NewMockWithShell(executor)

	err := git.Amend("new message")
	require.Error(t, err)
	assert.Equal(t, "amend error", err.Error())
}

func TestMock_AddAll_Success(t *testing.T) {
	executor := &executor.MockExecutor{}
	git := NewMockWithShell(executor)

	err := git.AddAll()
	require.NoError(t, err)
	assert.Contains(t, executor.Commands, "git add --all ")
}

func TestMock_AddAll_Error(t *testing.T) {
	executor := &executor.MockExecutor{Err: fmt.Errorf("add all error")}
	git := NewMockWithShell(executor)

	err := git.AddAll()
	require.Error(t, err)
	assert.Equal(t, "add all error", err.Error())
}

func TestMock_Reset_Success(t *testing.T) {
	executor := &executor.MockExecutor{}
	git := NewMockWithShell(executor)

	err := git.Reset("HEAD~1")
	require.NoError(t, err)
	assert.Contains(t, executor.Commands, "git reset --soft HEAD~1")
}

func TestMock_Reset_Error(t *testing.T) {
	executor := &executor.MockExecutor{Err: fmt.Errorf("reset error")}
	git := NewMockWithShell(executor)

	err := git.Reset("HEAD~1")
	require.Error(t, err)
	assert.Equal(t, "reset error", err.Error())
}

func TestMock_Append_ToCommit(t *testing.T) {
	git := NewMock()

	err := git.Append()
	require.NoError(t, err)
}

func TestMock_CommitMessage(t *testing.T) {
	git := NewMock()
	output, err := git.CommitMessage()

	require.NoError(t, err)
	expected := "feat(#42): current commit message"
	assert.Equal(t, expected, output)
}

func TestMock_CurrentBranch(t *testing.T) {
	git := NewMock()
	output, err := git.CurrentBranch()

	require.NoError(t, err)
	assert.Equal(t, "41_working_branch", output)
}

func TestMock_Diff(t *testing.T) {
	git := NewMock()
	output, err := git.Diff()

	require.NoError(t, err)
	assert.Equal(t, "mock-diff", output)
}

func TestMock_CurrentDiff(t *testing.T) {
	git := NewMock()

	output, err := git.CurrentDiff()

	require.NoError(t, err)
	assert.Equal(t, "current-mock-diff", output)
}

func TestMock_Remotes_WithoutShell(t *testing.T) {
	git := NewMockWithShell(nil)

	output, err := git.Remotes()

	require.NoError(t, err)
	first := "https://github.com/volodya-lombrozo/aidy.git"
	second := "https://github.com/volodya-lombrozo/forked-aidy.git"
	assert.Equal(t, first, output[0])
	assert.Equal(t, second, output[1])
}

func TestMock_Remotes_WithShell(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "https://github.com/vl/aidy.git"
	git := NewMockWithShell(shell)

	output, err := git.Remotes()

	require.NoError(t, err)
	assert.Equal(t, "https://github.com/vl/aidy.git", output[0])
}

func TestMock_Installed(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "git version 2.34.1"
	git := NewMockWithShell(shell)

	installed, err := git.Installed()

	require.NoError(t, err)
	assert.True(t, installed)
}

func TestMock_Checkout(t *testing.T) {
	git := NewMock()

	err := git.Checkout("main")

	require.NoError(t, err)
}

func TestMock_Tags(t *testing.T) {
	git := NewMock()

	output, err := git.Tags("does-not-matter")

	require.NoError(t, err)
	assert.Equal(t, []string{"v1.0", "v2.0"}, output)
}

func TestMock_WithError(t *testing.T) {
	git := NewMockWithError(fmt.Errorf("mock error"))

	_, err := git.Run("status")

	require.Error(t, err)
	assert.Equal(t, "mock error", err.Error())
}
