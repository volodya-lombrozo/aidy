package git

import (
	"github.com/volodya-lombrozo/aidy/executor"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealGit_AppendToCommit(t *testing.T) {
	mockExecutor := &executor.MockExecutor{}
	gitService := NewRealGit(mockExecutor, "")

	err := gitService.AppendToCommit()
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

func TestRealGetBranchName(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()
	gitService := NewRealGit(&executor.RealExecutor{}, dir)
	branchName, err := gitService.GetBranchName()
	if err != nil {
		t.Fatalf("Error getting branch name: %v", err)
	}
	if branchName != "main-branch" {
		t.Fatalf("Expected branch name 'main-branch', got '%s'", branchName)
	}
}

func TestRealGetBaseBranchName(t *testing.T) {
	repoDir, cleanup := setupTestRepo(t)
	defer cleanup()
	gitService := NewRealGit(&executor.RealExecutor{}, repoDir)
	baseBranch, err := gitService.GetBaseBranchName()
	if err != nil {
		t.Fatalf("Error getting base branch name: %v", err)
	}
	if baseBranch != "main" {
		t.Fatalf("Expected base branch name 'main', got '%s'", baseBranch)
	}
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
	gitService := NewRealGit(&executor.RealExecutor{}, repoDir)
	diff, err := gitService.GetDiff()
	if err != nil {
		t.Fatalf("Error getting diff: %v", err)
	}
	if diff == "" {
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
	gitService := NewRealGit(&executor.RealExecutor{}, repoDir)
	message, err := gitService.GetCurrentCommitMessage()
	if err != nil {
		t.Fatalf("Error getting current commit message: %v", err)
	}
	if message != commitMessage {
		t.Fatalf("Expected commit message '%s', got '%s'", commitMessage, message)
	}
}
