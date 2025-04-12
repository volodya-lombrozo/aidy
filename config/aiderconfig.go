package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type AiderConfig struct {
	Model        string `yaml:"model"`
	OpenaiApiKey string `yaml:"openai-api-key"`
}

func NewAiderConf(filepath string) *AiderConfig {
	configData, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	var config AiderConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		panic(err)
	}
	return &config
}

func (c *AiderConfig) GetOpenAIAPIKey() (string, error) {
	return c.OpenaiApiKey, nil
}

func (c *AiderConfig) GetGithubAPIKey() (string, error) {
	return "", nil
}

func (c *AiderConfig) GetModel() (string, error) {
	model := c.Model
	return model, nil
}
