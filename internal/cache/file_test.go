package cache

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileCache_SetsAndGets(t *testing.T) {
	tmp := temp(t)
	defer clean(t, tmp)
	c, err := NewFileCache(tmp)
	require.NoError(t, err, "Failed to create cache")

	_, ok := c.Get("missing")
	require.False(t, ok, "Expected missing key to return false")

	err = c.Set("foo", "bar")
	require.NoError(t, err, "Failed to set value in cache")

	val, ok := c.Get("foo")
	require.True(t, ok, "Expected key to be found")
	assert.Equal(t, "bar", val, "Expected value to be 'bar'")
}

func TestFileCache_FileCreateError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	tmp := temp(t)
	defer clean(t, tmp)

	chmod(t, tmp, 0000)
	defer chmod(t, tmp, 0755)

	_, err := NewFileCache(tmp)
	assert.Error(t, err, "Expected an error when the file cannot be created (read-only folder)")
}

func TestFileCache_FileWriteError(t *testing.T) {
	tmp := temp(t)
	defer clean(t, tmp)

	cache, err := NewFileCache(tmp)
	require.NoError(t, err, "Failed to create cache")

	chmod(t, tmp, 0444)
	defer chmod(t, tmp, 0644)

	err = cache.Set("key", "value")
	assert.Error(t, err, "Expected an error when the file cannot be written to")
}

func TestFileCache_Persistence(t *testing.T) {
	tmp := temp(t)
	defer clean(t, tmp)
	first, err := NewFileCache(tmp)
	require.NoError(t, err, "Failed to create cache")
	_ = first.Set("key1", "value1")

	reopened, err := NewFileCache(tmp)

	require.NoError(t, err, "Failed to reopen cache")
	val, ok := reopened.Get("key1")
	assert.True(t, ok, "Expected key to be found after reopening cache")
	assert.Equal(t, "value1", val, "Expected value to be 'value1'")
}

func TestFileCache_Overwrites(t *testing.T) {
	tmp := temp(t)
	defer clean(t, tmp)
	cache, _ := NewFileCache(tmp)

	_ = cache.Set("k", "v1")
	_ = cache.Set("k", "v2")
	val, _ := cache.Get("k")

	assert.Equal(t, "v2", val, "Expected value to be overwritten to 'v2'")
}

func temp(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "cache_test_*.json")
	require.NoError(t, err, "Failed to create temp file")
	cerr := f.Close()
	require.NoError(t, cerr, "Failed to close temp file")
	return f.Name()
}

func chmod(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	err := os.Chmod(path, mode)
	require.NoError(t, err, "Failed to change file permissions")
}

func clean(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove file %s: %v", path, err)
	}
}
