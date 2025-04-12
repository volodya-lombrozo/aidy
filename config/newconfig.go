package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type NewConfig struct {
	DefaultModel string                       `yaml:"default-model"`
	APIKeys      map[string]string            `yaml:"api-keys"`
	Models       map[string]map[string]string `yaml:"models"`
	Github       string                       `yaml:"github-api-key"`
}

func NewConf(filepath string) *NewConfig {
	configData, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	var config NewConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		panic(err)
	}
	return &config
}

func (c *NewConfig) GetOpenAIAPIKey() (string, error) {
	return c.APIKeys["openai"], nil
}

func (c *NewConfig) GetGithubAPIKey() (string, error) {
	return c.APIKeys["github"], nil
}

func (c *NewConfig) GetModel() (string, error) {
	model := c.DefaultModel
	return c.Models[model]["model-id"], nil
}
