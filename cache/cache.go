package cache

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Cache interface {
	Get(key string) (string, bool)
	Set(key, value string) error
}

type fileCache struct {
	mu    sync.RWMutex
	path  string
	store map[string]string
}

type gitCache struct {
	delegate Cache
	path     string
}

type mockCache struct {
}

func NewMockCache() Cache {
	return &mockCache{}
}

func (c *mockCache) Get(key string) (string, bool) {
	return "", true
}

func (c *mockCache) Set(key, value string) error {
	return nil
}

func NewFileCache(path string) (Cache, error) {
	c := &fileCache{path: path, store: map[string]string{}}
	f, err := os.Open(path)
	if err == nil {
		defer func() {
			if cerr := f.Close(); cerr != nil {
				err = cerr
			}
		}()
		_ = json.NewDecoder(f).Decode(&c.store)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return c, nil
}

func (c *fileCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[key]
	return val, ok
}

func (c *fileCache) Set(key, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
	return c.save()
}

func (c *fileCache) save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(c.path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = cerr
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c.store)
}

func NewGitMockCache() Cache {
	original := NewMockCache()
	return &gitCache{delegate: original}
}

func NewGitCache(path string) (Cache, error) {
	original, err := NewFileCache(path)
	if err != nil {
		return nil, err
	}
	return &gitCache{delegate: original, path: path}, nil
}

func (c *gitCache) Get(key string) (string, bool) {
	return c.delegate.Get(key)
}

func (c *gitCache) Set(key, value string) error {
	if err := ensureIgnored(c.path); err != nil {
		return err
	}
	return c.delegate.Set(key, value)
}

func ensureIgnored(filePath string) error {
	const entry = ".aidy"
	const gitignore = ".gitignore"
	file, err := os.Open(gitignore)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(gitignore, []byte(entry+"\n"), 0644)
		}
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			err = cerr
		}
	}()

	// Check if the entry is already present
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == entry {
			return nil // already present
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Append the entry
	f, err := os.OpenFile(gitignore, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = cerr
		}
	}()

	_, err = f.WriteString(entry + "\n")
	return err
}
