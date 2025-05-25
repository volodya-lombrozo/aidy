package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/volodya-lombrozo/aidy/internal/git"
)

type CascadeConfig struct {
	original Config
}

func NewCascade(gs git.Git) Config {
	original, ok := findAidyConf(gs)
	if !ok {
		original, ok = findAiderConf(gs)
	}
	if !ok {
		log.Fatalf("Can't find any configuration file")
	}
	return &CascadeConfig{original: original}
}

func (c *CascadeConfig) OpenAiKey() (string, error) {
	return c.original.OpenAiKey()
}
func (c *CascadeConfig) GithubKey() (string, error) {
	return c.original.GithubKey()
}
func (c *CascadeConfig) DeepseekKey() (string, error) {
	return c.original.DeepseekKey()
}
func (c *CascadeConfig) Model() (string, error) {
	return c.original.Model()
}

func findAidyConf(gs git.Git) (Config, bool) {
	all := possiblePaths(gs, ".aidy.conf")
	for _, p := range all {
		if exists(p) {
			conf, err := YamlConf(p)
			if err != nil {
				panic(fmt.Sprintf("Error reading config file %s: %v", p, err))
			}
			return conf, true
		}
	}
	return nil, false
}

func findAiderConf(gs git.Git) (Config, bool) {
	all := possiblePaths(gs, ".aider.conf")
	for _, p := range all {
		if exists(p) {
			return NewAider(p), true
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
