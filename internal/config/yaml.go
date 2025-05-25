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

func YamlConf(filepath string) (*YamlConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var config YamlConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *YamlConfig) OpenAiKey() (string, error) {
	return c.APIKeys["openai"], nil
}

func (c *YamlConfig) GithubKey() (string, error) {
	return c.APIKeys["github"], nil
}

func (c *YamlConfig) Model() (string, error) {
	model := c.DefaultModel
	return c.Models[model]["model-id"], nil
}

func (c *YamlConfig) DeepseekKey() (string, error) {
	return c.APIKeys["deepseek"], nil
}
