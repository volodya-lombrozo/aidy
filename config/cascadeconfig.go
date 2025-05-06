package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/volodya-lombrozo/aidy/git"
)

type CascadeConfig struct {
	original Config
}

func NewCascadeConfig(gs git.Git) Config {
	original, ok := findAidyConf(gs)
	if !ok {
		original, ok = findAiderConf(gs)
	}
	if !ok {
		log.Fatalf("Can't find any configuration file")
	}
	return &CascadeConfig{original: original}
}

func (c *CascadeConfig) GetOpenAIAPIKey() (string, error) {
	return c.original.GetOpenAIAPIKey()
}
func (c *CascadeConfig) GetGithubAPIKey() (string, error) {
	return c.original.GetGithubAPIKey()
}
func (c *CascadeConfig) GetDeepseekAPIKey() (string, error) {
	return c.original.GetDeepseekAPIKey()
}
func (c *CascadeConfig) GetModel() (string, error) {
	return c.original.GetModel()
}

func findAidyConf(gs git.Git) (Config, bool) {
	all := possiblePaths(gs, ".aidy.conf")
	for _, p := range all {
		if exists(p) {
			return YamlConf(p), true
		}
	}
	return nil, false
}

func findAiderConf(gs git.Git) (Config, bool) {
	all := possiblePaths(gs, ".aider.conf")
	for _, p := range all {
		if exists(p) {
			return NewAiderConf(p), true
		}
	}
	return nil, false
}

func possiblePaths(gs git.Git, filename string) []string {
	paths := []string{}
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, filename+".yaml"))
		paths = append(paths, filepath.Join(cwd, filename+".yml"))
	}
	if gitRoot, err := gs.Root(); err == nil {
		paths = append(paths, filepath.Join(gitRoot, filename+".yaml"))
		paths = append(paths, filepath.Join(gitRoot, filename+".yml"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, filename+".yaml"))
		paths = append(paths, filepath.Join(home, filename+".yml"))
	}
	return paths
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Fprintf(os.Stderr, "Error checking path %s: %v\n", path, err)
		return false
	}
	return true
}
