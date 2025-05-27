package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/git"
)

func TestNewGitCache_CreatesFileIfNotExists(t *testing.T) {
	tmp := t.TempDir()
	defer clean(t, tmp)
	name := "cache.json"
	file := filepath.Join(tmp, name)

	_, err := NewGitCache(name, git.NewMockWithDir(tmp))

	assert.NoError(t, err, "Expected no error when creating GitCache with a new file path")
	_, err = os.Stat(file)
	assert.NoError(t, err, "Expected file to be created")
}

func TestGitCache_RootError(t *testing.T) {
	gitserv := git.NewMockWithError(errors.New("failed to execute command"))

	_, err := NewGitCache("path", gitserv)

	assert.Error(t, err, "Expected an error when failing to get the root directory")
}

func TestGitCache_FileOpenError(t *testing.T) {
	mgit := git.NewMock()

	_, err := NewGitCache(`/\ / : * ? " < > |/existent/path`, mgit)

	assert.Error(t, err, "Expected an error when the file cannot be opened")
}

func TestNewGitCache_Get(t *testing.T) {
	tmp := temp(t)
	defer clean(t, tmp)
	c, err := NewGitCache(filepath.Base(tmp), git.NewMockWithDir(filepath.Dir(tmp)))
	require.NoError(t, err, "Failed to create GitCache")

	err = c.Set("testKey", "testValue")
	require.NoError(t, err, "Failed to set value in GitCache")

	val, ok := c.Get("testKey")
	assert.True(t, ok, "Expected key to be found")
	assert.Equal(t, "testValue", val, "Expected value to be 'testValue'")
}

func TestGitMockCache_CreatesGitignoreWithEntry(t *testing.T) {
	tmp := t.TempDir()
	defer clean(t, tmp)
	original, direrr := os.Getwd()
	require.NoError(t, direrr, "Failed to get current directory")
	defer func() { _ = os.Chdir(original) }()
	_ = os.Chdir(tmp)

	ch := NewGitMockCache(tmp)
	err := ch.Set("set", "mock value")

	assert.NoError(t, err)
	content, err := os.ReadFile(".gitignore")
	assert.NoError(t, err)
	assert.Contains(t, string(content), ".aidy")
}

func TestGitMockCache_AppendsIfMissing(t *testing.T) {
	tmp := t.TempDir()
	defer clean(t, tmp)
	original, direrr := os.Getwd()
	require.NoError(t, direrr, "Failed to get current directory")
	defer func() { _ = os.Chdir(original) }()
	_ = os.Chdir(tmp)
	_ = os.WriteFile(".gitignore", []byte("something-else\n"), 0644)

	ch := NewGitMockCache(tmp)
	err := ch.Set("set", "mock value")

	assert.NoError(t, err)
	content, err := os.ReadFile(".gitignore")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "something-else")
	assert.Contains(t, string(content), ".aidy")
}

func TestGitMockCache_DoesNothingIfAlreadyPresent(t *testing.T) {
	tmp := t.TempDir()
	defer clean(t, tmp)
	original, direrr := os.Getwd()
	require.NoError(t, direrr, "Failed to get current directory")
	defer func() { _ = os.Chdir(original) }()
	_ = os.Chdir(tmp)
	_ = os.WriteFile(".gitignore", []byte(".aidy\n"), 0644)

	ch := NewGitMockCache(tmp)
	err := ch.Set("set", "mock value")

	assert.NoError(t, err)
	content, err := os.ReadFile(".gitignore")
	assert.NoError(t, err)
	assert.Equal(t, ".aidy\n", string(content))
}
