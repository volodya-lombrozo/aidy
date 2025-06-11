package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type AiderConfig struct {
	ModelYaml        string `yaml:"model"`
	OpenaiApiKeyYaml string `yaml:"openai-api-key"`
}

func NewAider(filepath string) (*AiderConfig, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var config AiderConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *AiderConfig) GithubKey() (string, error) {
	return "", nil
}

func (c *AiderConfig) Model() (string, error) {
	return c.ModelYaml, nil
}

func (c *AiderConfig) Provider() (string, error) {
	return "openai", nil
}

func (c *AiderConfig) Token() (string, error) {
	return c.OpenaiApiKeyYaml, nil
}
