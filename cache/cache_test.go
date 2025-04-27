package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volodya-lombrozo/aidy/git"
)

func TestSetAndGet(t *testing.T) {
	fmt.Println("Start test")
	tmpFile := tempCacheFile(t)
	defer func() {
		if err := os.Remove(tmpFile); err != nil {
			t.Fatalf("Failed to remove temp file: %v", err)
		}
	}()

	c, err := NewFileCache(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	if _, ok := c.Get("missing"); ok {
		t.Errorf("Expected missing key to return false")
	}

	if err := c.Set("foo", "bar"); err != nil {
		t.Errorf("Set failed: %v", err)
	}

	if val, ok := c.Get("foo"); !ok || val != "bar" {
		t.Errorf("Expected 'bar', got '%s'", val)
	}
}
func TestNewGitCache_CreatesFileIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	name := "cache.json"
	tmpFile := filepath.Join(tmpDir, name)

	_, err := NewGitCache(name, git.NewMockGitWithDir(tmpDir))
	assert.NoError(t, err, "Expected no error when creating GitCache with a new file path")

	_, err = os.Stat(tmpFile)
	assert.NoError(t, err, "Expected file to be created")
}

func TestNewGitCache_Get(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer func() {
		if err := os.Remove(tmpFile); err != nil {
			t.Fatalf("Failed to remove temp file: %v", err)
		}
	}()
	c, err := NewGitCache(filepath.Base(tmpFile), git.NewMockGitWithDir(filepath.Dir(tmpFile)))
	if err != nil {
		t.Fatalf("Failed to create GitCache: %v", err)
	}
	err = c.Set("testKey", "testValue")
	if err != nil {
		t.Fatalf("Failed to set value in GitCache: %v", err)
	}

	val, ok := c.Get("testKey")

	assert.True(t, ok, "Expected key to be found")
	assert.Equal(t, "testValue", val, "Expected value to be 'testValue'")
}

func TestPersistence(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer func() {
		if err := os.Remove(tmpFile); err != nil {
			t.Fatalf("Failed to remove temp file: %v", err)
		}
	}()

	c1, err := NewFileCache(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	_ = c1.Set("key1", "value1")

	// Simulate closing and reopening
	c2, err := NewFileCache(tmpFile)
	if err != nil {
		t.Fatalf("Failed to reopen cache: %v", err)
	}

	if val, ok := c2.Get("key1"); !ok || val != "value1" {
		t.Errorf("Expected 'value1', got '%s'", val)
	}
}

func TestOverwrite(t *testing.T) {
	tmpFile := tempCacheFile(t)
	defer func() {
		if err := os.Remove(tmpFile); err != nil {
			t.Fatalf("Failed to remove temp file: %v", err)
		}
	}()

	c, _ := NewFileCache(tmpFile)
	_ = c.Set("k", "v1")
	_ = c.Set("k", "v2")

	if val, _ := c.Get("k"); val != "v2" {
		t.Errorf("Expected 'v2', got '%s'", val)
	}
}

func tempCacheFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "cache_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	} else {
		fmt.Printf("Created temp file %v\n", f.Name())
	}
	if cerr := f.Close(); cerr != nil {
		t.Fatalf("Failed to close temp file: %v", cerr)
	}
	return f.Name()
}

func TestEnsureIgnored_CreatesGitignoreWithEntry(t *testing.T) {
	tmp := t.TempDir()
	originalDir, direrr := os.Getwd()
	if direrr != nil {
		panic(direrr)
	}
	defer func() {
		_ = os.Chdir(originalDir)
		if err := os.RemoveAll(tmp); err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}()
	_ = os.Chdir(tmp)

	ch := NewGitMockCache(tmp)
	err := ch.Set("set", "mock value")

	assert.NoError(t, err)
	content, err := os.ReadFile(".gitignore")
	assert.NoError(t, err)
	assert.Contains(t, string(content), ".aidy")
}

func TestEnsureIgnored_AppendsIfMissing(t *testing.T) {
	tmp := t.TempDir()
	originalDir, direrr := os.Getwd()
	if direrr != nil {
		panic(direrr)
	}
	defer func() {
		_ = os.Chdir(originalDir)
		if err := os.RemoveAll(tmp); err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}()
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

func TestEnsureIgnored_DoesNothingIfAlreadyPresent(t *testing.T) {
	tmp := t.TempDir()
	originalDir, direrr := os.Getwd()
	if direrr != nil {
		panic(direrr)
	}
	defer func() {
		_ = os.Chdir(originalDir)
		if err := os.RemoveAll(tmp); err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}()
	_ = os.Chdir(tmp)
	_ = os.WriteFile(".gitignore", []byte(".aidy\n"), 0644)

	ch := NewGitMockCache(tmp)
	err := ch.Set("set", "mock value")

	assert.NoError(t, err)
	content, err := os.ReadFile(".gitignore")
	assert.NoError(t, err)
	assert.Equal(t, ".aidy\n", string(content)) // nothing added
}
