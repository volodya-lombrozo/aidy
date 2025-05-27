package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type fileCache struct {
	mu    sync.RWMutex
	path  string
	store map[string]string
}

func NewFileCache(path string) (Cache, error) {
	path = filepath.FromSlash(path)
	fcache := &fileCache{path: path, store: map[string]string{}}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			dir := filepath.Dir(path)
			if mkerr := os.MkdirAll(dir, 0755); mkerr != nil {
				return nil, mkerr
			}
			created, cerr := os.Create(path)
			if cerr != nil {
				return nil, cerr
			}
			if cerr := created.Close(); cerr != nil {
				return nil, cerr
			}
		} else {
			return nil, err
		}
	} else {
		_ = json.NewDecoder(file).Decode(&fcache.store)
		if cerr := file.Close(); cerr != nil {
			return nil, cerr
		}
	}
	return fcache, nil
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
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(c.store)
	if err != nil {
		return err
	}
	return f.Close()
}
