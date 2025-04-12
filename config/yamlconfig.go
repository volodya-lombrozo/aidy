package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type YAMLConfig struct {
	OpenAIAPIKey string `yaml:"openai-api-key"`
	GitHubAPIKey string `yaml:"github-api-key"`
}

func (c *YAMLConfig) GetModel() (string, error) {
	return "gpt-4o", nil
}

func NewYAMLConfig(filepath string) (*YAMLConfig, error) {
	configData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}
	var config YAMLConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}
	return &config, nil
}

func (c *YAMLConfig) GetOpenAIAPIKey() (string, error) {
	if c.OpenAIAPIKey == "" {
		return "", fmt.Errorf("OpenAI API key not found in config file")
	}
	return c.OpenAIAPIKey, nil
}

func (c *YAMLConfig) GetGithubAPIKey() (string, error) {
	if c.GitHubAPIKey == "" {
		return "", fmt.Errorf("OpenAI API key not found in config file")
	}
	return c.GitHubAPIKey, nil
}
