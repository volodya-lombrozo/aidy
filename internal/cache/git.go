package cache

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/volodya-lombrozo/aidy/internal/git"
)

type gitCache struct {
	delegate Cache
	gs       git.Git
}

func NewGitCache(path string, gs git.Git) (Cache, error) {
	root, err := gs.Root()
	if err != nil {
		return nil, err
	}
	original, err := NewFileCache(filepath.Join(root, path))
	if err != nil {
		return nil, err
	}
	return &gitCache{delegate: original, gs: gs}, nil
}

func NewGitMockCache(gitdir string) Cache {
	return &gitCache{delegate: NewMockCache(), gs: git.NewMockWithDir(gitdir)}
}

func (c *gitCache) Get(key string) (string, bool) {
	return c.delegate.Get(key)
}

func (c *gitCache) Set(key, value string) error {
	if err := ensureIgnored(c.gs); err != nil {
		return err
	}
	return c.delegate.Set(key, value)
}

func ensureIgnored(gs git.Git) error {
	const entry = ".aidy"
	root, _ := gs.Root()
	gitignore := filepath.Join(root, ".gitignore")
	file, err := os.Open(gitignore)
	defer fclose(file)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(gitignore, []byte(entry+"\n"), 0644)
		}
		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == entry {
			return nil
		}
	}
	err = scanner.Err()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(gitignore, os.O_APPEND|os.O_WRONLY, 0644)
	defer fclose(f)
	if err != nil {
		return err
	}
	_, err = f.WriteString(entry + "\n")
	return err
}

func fclose(file *os.File) {
	if file == nil {
		return
	}
	err := file.Close()
	if err != nil {
		panic("Failed to close file: " + err.Error())
	}
}
