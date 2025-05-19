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
	"github.com/volodya-lombrozo/aidy/executor"
)

func TestRealGit_Run_Successful(t *testing.T) {
	mock := &executor.MockExecutor{
		Output: "success",
		Err:    nil,
	}
	service := NewGit(mock)
	output, err := service.Run("status")
	require.NoError(t, err)
	assert.Equal(t, "success", output)
}

func TestRealGit_Run_Failure(t *testing.T) {
	mock := &executor.MockExecutor{
		Output: "",
		Err:    fmt.Errorf("error"),
	}
	service := NewGit(mock)
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
	service := NewGit(mock)

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
	gs := NewGit(shell)

	tags, err := gs.Tags()

	require.NoError(t, err)
	expectedTags := []string{"v1.0.0", "v1.1.0", "v2.1.0"}
	assert.Equal(t, expectedTags, tags)
}

func TestRealGit_Tags_FetchError(t *testing.T) {
	shell := &executor.MockExecutor{
		Err: fmt.Errorf("fetch error"),
	}
	gs := NewGit(shell)

	tags, err := gs.Tags()

	require.Error(t, err)
	assert.Nil(t, tags)
	assert.Contains(t, err.Error(), "error fetching tags")
}

func TestRealGit_Tags_ListError(t *testing.T) {
	shell := &executor.MockExecutor{
		Output: "",
		Err:    fmt.Errorf("list error"),
	}
	gs := NewGit(shell)

	tags, err := gs.Tags()

	require.Error(t, err)
	assert.Nil(t, tags)
	assert.Contains(t, err.Error(), "list error")
}

func TestRealGit_Checkout_Success(t *testing.T) {
	shell := &executor.MockExecutor{}
	gs := NewGit(shell)
	branch := "feature-branch"

	err := gs.Checkout(branch)

	require.NoError(t, err)
	expectedCommand := "git checkout -b " + branch
	assert.Equal(t, len(shell.Commands), 1, "Expected 1 command to be executed")
	assert.Contains(t, shell.Commands[0], expectedCommand)
}

func TestRealGit_Checkout_Failure(t *testing.T) {
	shell := &executor.MockExecutor{
		Err: fmt.Errorf("checkout error"),
	}
	gitService := NewGit(shell)
	branchName := "feature-branch"

	err := gitService.Checkout(branchName)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "checkout error")
}

func TestRealGit_Amend(t *testing.T) {
	mockExecutor := &executor.MockExecutor{}
	gitService := NewGit(mockExecutor, "")

	message := "Updated commit message"
	err := gitService.Amend(message)
	require.NoError(t, err)

	expectedCommand := "git commit --amend -m " + message
	if len(mockExecutor.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(mockExecutor.Commands))
	}

	if !strings.Contains(mockExecutor.Commands[0], expectedCommand) {
		t.Errorf("Expected command '%s', got '%s'", expectedCommand, mockExecutor.Commands[0])
	}
}

func TestRealGit_AddAll(t *testing.T) {
	mockExecutor := &executor.MockExecutor{}
	gitService := NewGit(mockExecutor, "")

	err := gitService.AddAll()
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
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()

	gitService := NewGit(executor.NewRealExecutor(), repoDir)

	filePath := filepath.Join(repoDir, "resetfile.txt")
	if err := os.WriteFile(filePath, []byte("Reset content"), 0644); err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running command: %v", err)
	}
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
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()
	gitService := NewGit(executor.NewRealExecutor(), repoDir)

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
			result, err := NewGit(mock).Remotes()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRealGit_AppendToCommit(t *testing.T) {
	mockExecutor := &executor.MockExecutor{}
	gitService := NewGit(mockExecutor, "")

	err := gitService.Append()
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
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	gs := NewGit(executor.NewRealExecutor(), dir)

	branch, err := gs.CurrentBranch()

	require.NoError(t, err, "Expected no error during branch retrieval")
	assert.Equal(t, "main-branch", branch, "Expected branch name to be 'main-branch'")
}

func TestRealGetBaseBranchName(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()
	gs := NewGit(executor.NewRealExecutor(), repoDir)

	base, err := gs.BaseBranch()

	require.NoError(t, err, "Expected no error during base branch retrieval")
	assert.Equal(t, "main", base, "Expected base branch name to be 'main'")
}

func TestRealGetDiff(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
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
	gitService := NewGit(executor.NewRealExecutor(), repoDir)
	diff, err := gitService.Diff()
	if err != nil {
		t.Fatalf("Error getting diff: %v", err)
	}
	if diff == "" {
		t.Fatal("Expected non-empty diff")
	}
}

func TestRealGetCurrentDiff(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
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
	gitService := NewGit(executor.NewRealExecutor(), repoDir)
	diff, err := gitService.Diff()
	if err != nil {
		t.Fatalf("Error getting diff: %v", err)
	}
	if diff == "Hello, Git!" {
		t.Fatal("Expected non-empty diff")
	}
}

func TestRealGetCurrentCommitMessage(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
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
	gitService := NewGit(executor.NewRealExecutor(), repoDir)
	message, err := gitService.CommitMessage()
	if err != nil {
		t.Fatalf("Error getting current commit message: %v", err)
	}
	if message != commitMessage {
		t.Fatalf("Expected commit message '%s', got '%s'", commitMessage, message)
	}
}

func TestRealGitInstalled(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()
	gitService := NewGit(executor.NewRealExecutor(), repoDir)

	installed, err := gitService.Installed()

	require.NoError(t, err)
	assert.True(t, installed)
}

func setupTestRepo(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "testrepo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	cmd := exec.Command("git", "init", "--initial-branch", "main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	_ = cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial commit", "--allow-empty")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to make an initial commit: %v", err)
	}
	cmd = exec.Command("git", "checkout", "-b", "main-branch")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create 'main-branch' branch: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "second commit", "--allow-empty")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to make an initial commit: %v", err)
	}
	return tempDir, func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Error removing temp directory: %v", err)
		}
	}
}
