package cache

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/git"
)

func repository(t *testing.T) (string, git.Git) {
	dir := t.TempDir()
	init := exec.Command("git", "init", "--quiet")
	init.Dir = dir
	require.NoError(t, init.Run(), "failed to initialize git repository")
	gs, err := git.NewGit(executor.NewReal(), dir)
	require.NoError(t, err, "failed to create git")
	return dir, gs
}

func TestLastCommand_WithLast_KeepsCommandInsideGitDir(t *testing.T) {
	dir, gs := repository(t)
	last := NewLastCommand(gs)

	last.WithLast("gh pr create\n  --title \"title\"")

	saved, err := os.ReadFile(filepath.Join(dir, ".git", "aidy-last"))
	require.NoError(t, err, "expected the last command to be kept in .git")
	assert.Equal(t, "gh pr create\n  --title \"title\"", string(saved), "expected the command to be saved as is")
	_, err = os.Stat(filepath.Join(dir, ".aidy"))
	assert.True(t, os.IsNotExist(err), "expected no .aidy directory to be created")
}

func TestLastCommand_Last_ReturnsSavedCommand(t *testing.T) {
	_, gs := repository(t)
	last := NewLastCommand(gs)
	last.WithLast("gh issue create")

	command, err := last.Last()

	require.NoError(t, err, "expected no error when reading the last command")
	assert.Equal(t, "gh issue create", command, "expected the saved command back")
}

func TestLastCommand_Last_EmptyWhenNothingSaved(t *testing.T) {
	_, gs := repository(t)

	command, err := NewLastCommand(gs).Last()

	require.NoError(t, err, "expected no error when nothing was saved")
	assert.Equal(t, "", command, "expected no last command")
}

func TestLastCommand_Last_FailsOutsideRepository(t *testing.T) {
	last := NewLastCommand(git.NewMockWithError(errors.New("not a git repository")))

	_, err := last.Last()

	assert.Error(t, err, "expected an error when the git dir can't be found")
}

func TestLastCommand_WithLast_PanicsOutsideRepository(t *testing.T) {
	last := NewLastCommand(git.NewMockWithError(errors.New("not a git repository")))

	assert.Panics(t, func() {
		last.WithLast("gh pr create")
	}, "expected a panic when the git dir can't be found")
}

func TestMockLastCommand_WithLast(t *testing.T) {
	last := NewMockLastCommand()

	last.WithLast("gh issue create")

	command, err := last.Last()
	require.NoError(t, err)
	assert.Equal(t, "gh issue create", command, "expected the mock to keep the command")
}
