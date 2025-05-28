package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/executor"
)

func TestRealGit_Run_Successful(t *testing.T) {
	mock := &executor.MockExecutor{
		Output: "success",
		Err:    nil,
	}
	service, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")
	output, err := service.Run("status")
	require.NoError(t, err)
	assert.Equal(t, "success", output)
}

func TestRealGit_Run_Failure(t *testing.T) {
	mock := &executor.MockExecutor{
		Output: "",
		Err:    fmt.Errorf("error"),
	}
	service, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")
	output, err := service.Run("status")
	require.Error(t, err)
	assert.Equal(t, "", output)
	assert.Contains(t, err.Error(), "error running git command")
}

func TestRealGit_Remotes(t *testing.T) {
	mock := &executor.MockExecutor{
		Output: "origin\thttps://github.com/user/repo.git (fetch)\norigin\thttps://github.com/user/repo.git (push)\nupstream\thttps://github.com/another/repo.git (fetch)\nupstream\thttps://github.com/another/repo.git (push)\n",
		Err:    nil,
	}
	service, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")

	urls, err := service.Remotes()

	require.NoError(t, err)
	expected := []string{
		"https://github.com/user/repo.git",
		"https://github.com/another/repo.git",
	}
	assert.Equal(t, urls, expected)
}

func TestRealGit_Tags_Success(t *testing.T) {
	shell := &executor.MockExecutor{
		Output: "v1.0.0\nv1.1.0\nv2.1.0\n",
		Err:    nil,
	}
	gs, err := NewGit(shell)
	require.NoError(t, err, "git should be createad without any problems")

	tags, err := gs.Tags("upstream")

	require.NoError(t, err)
	expectedTags := []string{"v1.0.0", "v1.1.0", "v2.1.0"}
	assert.Equal(t, expectedTags, tags)
}

func TestRealGit_Tags_FetchError(t *testing.T) {
	shell := &executor.MockExecutor{
		Err: fmt.Errorf("fetch error"),
	}
	gs, err := NewGit(shell)
	require.NoError(t, err, "git should be createad without any problems")

	tags, err := gs.Tags("errepo")

	require.Error(t, err)
	assert.Nil(t, tags)
	assert.Contains(t, err.Error(), "error fetching tags")
}

func TestRealGit_Tags_ListError(t *testing.T) {
	shell := &executor.MockExecutor{
		Output: "",
		Err:    fmt.Errorf("list error"),
	}
	gs, err := NewGit(shell)
	require.NoError(t, err, "git should be createad without any problems")

	tags, err := gs.Tags("origin")

	require.Error(t, err)
	assert.Nil(t, tags)
	assert.Contains(t, err.Error(), "list error")
}

func TestRealGit_Checkout_Success(t *testing.T) {
	shell := &executor.MockExecutor{}
	gs, err := NewGit(shell)
	require.NoError(t, err, "git should be createad without any problems")
	branch := "feature-branch"

	err = gs.Checkout(branch)

	require.NoError(t, err)
	expectedCommand := "git checkout -b " + branch
	assert.Equal(t, len(shell.Commands), 1, "Expected 1 command to be executed")
	assert.Contains(t, shell.Commands[0], expectedCommand)
}

func TestRealGit_Checkout_Failure(t *testing.T) {
	shell := &executor.MockExecutor{
		Err: fmt.Errorf("checkout error"),
	}
	git, err := NewGit(shell)
	require.NoError(t, err, "git should be createad without any problems")
	branch := "feature-branch"

	err = git.Checkout(branch)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "checkout error")
}

func TestRealGit_Amend(t *testing.T) {
	mockExecutor := &executor.MockExecutor{}
	git, err := NewGit(mockExecutor, "")
	require.NoError(t, err, "git should be createad without any problems")

	message := "Updated commit message"
	err = git.Amend(message)
	require.NoError(t, err)

	expectedCommand := "git commit --amend -m " + message
	require.Len(t, mockExecutor.Commands, 1, "Expected 1 command")
	assert.Contains(t, mockExecutor.Commands[0], expectedCommand, "Expected command to be executed")
}

func TestRealGit_AddAll(t *testing.T) {
	mockExecutor := &executor.MockExecutor{}
	git, err := NewGit(mockExecutor, "")
	require.NoError(t, err, "git should be createad without any problems")

	err = git.AddAll()
	require.NoError(t, err)

	expectedCommand := "git add --all"
	if len(mockExecutor.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(mockExecutor.Commands))
	}

	if !strings.Contains(mockExecutor.Commands[0], expectedCommand) {
		t.Errorf("Expected command '%s', got '%s'", expectedCommand, mockExecutor.Commands[0])
	}
}

func TestRealGit_Reset(t *testing.T) {
	repoDir, cleanup := setup(t)
	defer cleanup()

	gitService, err := NewGit(executor.NewReal(), repoDir)
	require.NoError(t, err, "git should be createad without any problems")

	filePath := filepath.Join(repoDir, "resetfile.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("Reset content"), 0644), "Error writing to file")
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	require.NoError(t, cmd.Run(), "Error running command")
	cmd = exec.Command("git", "commit", "-m", "Commit for reset test")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running command: %v", err)
	}

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	commitHash, err := cmd.Output()
	if err != nil {
		t.Fatalf("Error getting commit hash: %v", err)
	}
	hash := strings.TrimSpace(string(commitHash))

	err = gitService.Reset(hash)

	require.NoError(t, err)
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoDir
	statusOutput, err := cmd.Output()
	if err != nil {
		t.Fatalf("Error getting git status: %v", err)
	}
	assert.Empty(t, strings.TrimSpace(string(statusOutput)), "Expected working directory to be clean after reset")
}

func TestRealGitRoot(t *testing.T) {
	repoDir, cleanup := setup(t)
	defer cleanup()
	gitService, err := NewGit(executor.NewReal(), repoDir)
	require.NoError(t, err, "git should be createad without any problems")

	root, err := gitService.Root()

	require.NoError(t, err)
	expectedRoot, err := filepath.EvalSymlinks(strings.TrimSpace(repoDir))
	require.NoError(t, err)
	assert.Equal(t, filepath.ToSlash(expectedRoot), filepath.ToSlash(strings.TrimSpace(root)))
}

func TestRealGit_Remotes_Parametrised(t *testing.T) {
	tests := []struct {
		remotes  string
		expected []string
	}{
		{"origin\thttps://github.com/u/r.git (fetch)", []string{"https://github.com/u/r.git"}},
		{"remote\thttps://github.com/u/r.git (fetch)", []string{"https://github.com/u/r.git"}},
		{"upstream\thttps://github.com/u/r.git (fetch)", []string{"https://github.com/u/r.git"}},
		{"upstream\thttps://github.com/u/r.git (fetch)\norigin\thttps://github.com/u/r.git (fetch)", []string{"https://github.com/u/r.git"}},
	}
	for _, tc := range tests {
		t.Run(tc.remotes, func(t *testing.T) {
			mock := &executor.MockExecutor{Output: tc.remotes, Err: nil}
			git, err := NewGit(mock)
			require.NoError(t, err, "git should be created without any problems")
			result, err := git.Remotes()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRealGit_AppendToCommit(t *testing.T) {
	mockExecutor := &executor.MockExecutor{}
	gitService, err := NewGit(mockExecutor, "")
	require.NoError(t, err, "git should be createad without any problems")

	err = gitService.Append()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedCommands := []string{
		"git add --all",
		"git commit --amend --no-edit",
	}

	if len(mockExecutor.Commands) != len(expectedCommands) {
		t.Fatalf("Expected %d commands, got %d", len(expectedCommands), len(mockExecutor.Commands))
	}

	for i, cmd := range expectedCommands {
		if !strings.Contains(mockExecutor.Commands[i], cmd) {
			t.Errorf("Expected command '%s', got '%s'", cmd, mockExecutor.Commands[i])
		}
	}
}

func TestRealGetBranchName(t *testing.T) {
	dir, cleanup := setup(t)
	defer cleanup()
	gs, err := NewGit(executor.NewReal(), dir)
	require.NoError(t, err, "git should be createad without any problems")

	branch, err := gs.CurrentBranch()

	require.NoError(t, err, "Expected no error during branch retrieval")
	assert.Equal(t, "main-branch", branch, "Expected branch name to be 'main-branch'")
}

func TestRealGetBaseBranchName(t *testing.T) {
	repoDir, cleanup := setup(t)
	defer cleanup()
	gs, err := NewGit(executor.NewReal(), repoDir)
	require.NoError(t, err, "git should be createad without any problems")

	base, err := gs.BaseBranch()

	require.NoError(t, err, "Expected no error during base branch retrieval")
	assert.Equal(t, "main", base, "Expected base branch name to be 'main'")
}

func TestRealGetDiff(t *testing.T) {
	repoDir, cleanup := setup(t)
	defer cleanup()
	filePath := filepath.Join(repoDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running command: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "Add hello world")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running command: %v", err)
	}
	require.NoError(t, os.WriteFile(filePath, []byte("Hello, Git!"), 0644), "Error writing to file")
	gitService, err := NewGit(executor.NewReal(), repoDir)
	require.NoError(t, err, "git should be createad without any problems")
	diff, err := gitService.Diff()
	require.NoError(t, err, "Error getting diff")
	assert.NotEmpty(t, diff, "Expected non-empty diff")
}

func TestRealGetCurrentDiff(t *testing.T) {
	repoDir, cleanup := setup(t)
	defer cleanup()
	filePath := filepath.Join(repoDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running command: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "Add hello world")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running command: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("Hello, Git!"), 0644); err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}
	gitService, err := NewGit(executor.NewReal(), repoDir)
	require.NoError(t, err, "git should be createad without any problems")
	diff, err := gitService.Diff()
	if err != nil {
		t.Fatalf("Error getting diff: %v", err)
	}
	if diff == "Hello, Git!" {
		t.Fatal("Expected non-empty diff")
	}
}

func TestRealGetCurrentCommitMessage(t *testing.T) {
	repoDir, cleanup := setup(t)
	defer cleanup()
	// Create a new commit to test
	filePath := filepath.Join(repoDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("Hello, Commit!"), 0644); err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running command: %v", err)
	}
	commitMessage := "Test commit message"
	cmd = exec.Command("git", "commit", "-m", commitMessage)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running command: %v", err)
	}
	gitService, err := NewGit(executor.NewReal(), repoDir)
	require.NoError(t, err, "git should be createad without any problems")
	message, err := gitService.CommitMessage()
	if err != nil {
		t.Fatalf("Error getting current commit message: %v", err)
	}
	if message != commitMessage {
		t.Fatalf("Expected commit message '%s', got '%s'", commitMessage, message)
	}
}

func TestRealGitInstalled(t *testing.T) {
	repoDir, cleanup := setup(t)
	defer cleanup()
	gitService, err := NewGit(executor.NewReal(), repoDir)
	require.NoError(t, err, "git should be createad without any problems")

	installed, err := gitService.Installed()

	require.NoError(t, err)
	assert.True(t, installed)
}

func TestRealGit_CantCreate(t *testing.T) {
	fallback := func() (string, error) {
		return "", fmt.Errorf("failed to get current working directory")
	}
	_, err := NewGitFallback(executor.NewReal(), fallback)
	require.Error(t, err, "Expected error when creating git service with non-existent directory")
	assert.Contains(t, err.Error(), "failed to get current working directory")
}

func TestRealGit_CantAmmend(t *testing.T) {
	mock := &executor.MockExecutor{
		Err: fmt.Errorf("amend error"),
	}
	git, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")
	err = git.Amend("New commit message")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error amending commit")
}

func TestRealGit_CantReset(t *testing.T) {
	mock := &executor.MockExecutor{
		Err: fmt.Errorf("reset error"),
	}
	git, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")
	err = git.Reset("HEAD~1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error resetting to HEAD~1")
}

func TestRealGit_CantRetriveCurrentBranch(t *testing.T) {
	mock := &executor.MockExecutor{
		Err: fmt.Errorf("current branch error"),
	}
	git, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")
	branch, err := git.CurrentBranch()
	require.Error(t, err)
	assert.Empty(t, branch)
	assert.Contains(t, err.Error(), "current branch error")
}

func TestRealGit_CantAppend(t *testing.T) {
	mock := &executor.MockExecutor{
		Err: fmt.Errorf("append error"),
	}
	git, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")
	err = git.Append()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error adding changes")
}

func TestRealGit_CantRetrieveBaseBranch(t *testing.T) {
	mock := &executor.MockExecutor{
		Err: fmt.Errorf("base branch error"),
	}
	git, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")
	base, err := git.BaseBranch()
	require.Error(t, err)
	assert.Empty(t, base)
	assert.Contains(t, err.Error(), "neither 'main' nor 'master' branch exists")
}

func TestRealGit_CantRetrieveDiff(t *testing.T) {
	mock := &executor.MockExecutor{
		Err: fmt.Errorf("diff error"),
	}
	git, err := NewGit(mock)
	require.NoError(t, err, "git should be createad without any problems")

	diff, err := git.Diff()

	require.Error(t, err)
	assert.Empty(t, diff)
	assert.Contains(t, err.Error(), "neither 'main' nor 'master' branch exists")
}

func TestRealGit_RetrieveCurrentDiff(t *testing.T) {
	repo, cleanup := setup(t)
	defer cleanup()
	err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("Hello, World!"), 0644)
	require.NoError(t, err, "Error writing to file")
	git, err := NewGit(executor.NewReal(), repo)
	require.NoError(t, err, "git should be createad without any problems")
	err = git.AddAll()
	require.NoError(t, err, "Expected no error during adding all changes")

	diff, err := git.CurrentDiff()

	require.NoError(t, err, "Expected no error during current diff retrieval")
	assert.NotEmpty(t, diff, "Expected non-empty current diff")
	assert.Contains(t, diff, "Hello, World!", "Expected diff to contain 'Hello, World!'")
}

func TestRealGit_ReadsLogs(t *testing.T) {
	repo, cleanup := setup(t)
	defer cleanup()
	git, err := NewGit(executor.NewReal(), repo)
	require.NoError(t, err, "git should be createad without any problems")

	logs, err := git.Log("HEAD~1")

	require.NoError(t, err, "Expected no error during log retrieval")
	assert.NotEmpty(t, logs, "Expected non-empty logs")
	assert.Contains(t, logs[0], "second commit", "Expected log to contain 'second commit'")
}

func TestRealGit_AddAll_Failure(t *testing.T) {
	shell := &executor.MockExecutor{
		Err: fmt.Errorf("add all error"),
	}
	git, err := NewGit(shell, "")
	require.NoError(t, err, "git should be created without any problems")

	err = git.AddAll()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error adding all changes")
}

func TestRealGit_Installed_Failure(t *testing.T) {
	shell := &executor.MockExecutor{
		Err: fmt.Errorf("version error"),
	}
	git, err := NewGit(shell, "")
	require.NoError(t, err, "git should be created without any problems")

	installed, err := git.Installed()

	require.Error(t, err)
	assert.False(t, installed)
	assert.Contains(t, err.Error(), "version error")
}

func TestRealGit_Root_Failure(t *testing.T) {
	shell := &executor.MockExecutor{
		Err: fmt.Errorf("root error"),
	}
	git, err := NewGit(shell, "")
	require.NoError(t, err, "git should be created without any problems")

	root, err := git.Root()

	require.Error(t, err)
	assert.Empty(t, root)
	assert.Contains(t, err.Error(), "root error")
}

func TestRealGit_Log_Failure(t *testing.T) {
	mockExecutor := &executor.MockExecutor{
		Err: fmt.Errorf("log error"),
	}
	git, err := NewGit(mockExecutor, "")
	require.NoError(t, err, "git should be created without any problems")

	logs, err := git.Log("HEAD~1")
	require.Error(t, err)
	assert.Nil(t, logs)
	assert.Contains(t, err.Error(), "log error")
}

func TestRealGit_Log_All(t *testing.T) {
	tmp, cleanup := setup(t)
	defer cleanup()
	git, err := NewGit(executor.NewReal(), tmp)
	require.NoError(t, err, "git should be created without any problems")

	logs, err := git.Log("")

	require.NoError(t, err, "Expected no error during log retrieval")
	assert.NotEmpty(t, logs, "Expected non-empty logs")
	assert.Len(t, logs, 2, "Expected exactly 2 log entries")
	assert.Contains(t, logs[0], "second commit", "Expected log to contain 'second commit'")
}

func setup(t *testing.T) (string, func()) {
	tmp, err := os.MkdirTemp("", "testrepo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	cmd := exec.Command("git", "init", "--initial-branch", "main")
	cmd.Dir = tmp
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmp
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmp
	_ = cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial commit", "--allow-empty")
	cmd.Dir = tmp
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to make an initial commit: %v", err)
	}
	cmd = exec.Command("git", "checkout", "-b", "main-branch")
	cmd.Dir = tmp
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create 'main-branch' branch: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "second commit", "--allow-empty")
	cmd.Dir = tmp
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to make an initial commit: %v", err)
	}
	return tmp, func() {
		require.NoError(t, os.RemoveAll(tmp), "Error removing temp directory")
	}
}
