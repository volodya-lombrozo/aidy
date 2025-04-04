package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type YAMLConfig struct {
	OpenAIAPIKey string `yaml:"openai-api-key"`
}

func NewYAMLConfig(filePath string) (*YAMLConfig, error) {
	configData, err := os.ReadFile(filePath)
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
