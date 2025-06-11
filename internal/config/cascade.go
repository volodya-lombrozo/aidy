package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/volodya-lombrozo/aidy/internal/git"
)

type CascadeConfig struct {
	original Config
}

func NewCascade(gs git.Git) (Config, error) {
	return NewCascadeInDirs(os.Getwd, gs.Root, os.UserHomeDir)
}

func NewCascadeInDirs(folders ...func() (string, error)) (Config, error) {
	original, err := findAidyConf(folders...)
	if err != nil {
		original, err = findAiderConf(folders...)
	}
	if err != nil {
		return nil, fmt.Errorf("can't find any configuration file")
	}
	return &CascadeConfig{original: original}, nil
}

func (c *CascadeConfig) GithubKey() (string, error) {
	return c.original.GithubKey()
}

func (c *CascadeConfig) Model() (string, error) {
	return c.original.Model()
}

func (c *CascadeConfig) Provider() (string, error) {
	return c.original.Provider()
}

func (c *CascadeConfig) Token() (string, error) {
	return c.original.Token()
}

func findAidyConf(folders ...func() (string, error)) (Config, error) {
	all := possibleFiles(".aidy.conf", folders...)
	for _, p := range all {
		if exists(p) {
			conf, err := YamlConf(p)
			if err != nil {
				return nil, fmt.Errorf("error reading config file %s: %v", p, err)
			}
			return conf, nil
		}
	}
	return nil, fmt.Errorf("no .aidy.conf found in any of the expected locations")
}

func findAiderConf(folders ...func() (string, error)) (Config, error) {
	all := possibleFiles(".aider.conf", folders...)
	for _, p := range all {
		if exists(p) {
			conf, err := NewAider(p)
			if err != nil {
				return nil, fmt.Errorf("error reading config file %s: %v", p, err)
			}
			return conf, nil
		}
	}
	return nil, fmt.Errorf("no .aider.conf found in any of the expected locations")
}

func possibleFiles(filename string, locations ...func() (string, error)) []string {
	paths := []string{}
	for _, location := range locations {
		if loc, err := location(); err == nil {
			paths = append(paths, filepath.Join(loc, filename+".yaml"))
			paths = append(paths, filepath.Join(loc, filename+".yml"))
		}
	}
	return paths
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}
