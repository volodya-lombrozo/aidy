package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type YamlConfig struct {
	DefaultModel string                       `yaml:"default-model"`
	APIKeys      map[string]string            `yaml:"api-keys"`
	Models       map[string]map[string]string `yaml:"models"`
	Github       string                       `yaml:"github-api-key"`
}

func YamlConf(filepath string) *YamlConfig {
	configData, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	var config YamlConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		panic(err)
	}
	return &config
}

func (c *YamlConfig) GetOpenAIAPIKey() (string, error) {
	return c.APIKeys["openai"], nil
}

func (c *YamlConfig) GetGithubAPIKey() (string, error) {
	return c.APIKeys["github"], nil
}

func (c *YamlConfig) GetModel() (string, error) {
	model := c.DefaultModel
	return c.Models[model]["model-id"], nil
}

func (c *YamlConfig) GetDeepseekAPIKey() (string, error) {
	return c.APIKeys["deepseek"], nil
}
