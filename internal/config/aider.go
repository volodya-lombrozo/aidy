package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type AiderConfig struct {
	ModelYaml        string `yaml:"model"`
	OpenaiApiKeyYaml string `yaml:"openai-api-key"`
}

func NewAider(filepath string) *AiderConfig {
	file, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	var config AiderConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}
	return &config
}

func (c *AiderConfig) OpenAiKey() (string, error) {
	return c.OpenaiApiKeyYaml, nil
}

func (c *AiderConfig) GithubKey() (string, error) {
	return "", nil
}

func (c *AiderConfig) Model() (string, error) {
	return c.ModelYaml, nil
}

func (c *AiderConfig) DeepseekKey() (string, error) {
	return "unknown", nil
}
