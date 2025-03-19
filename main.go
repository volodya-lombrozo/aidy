package main

import (
    "fmt"
    "log"
    "github.com/volodya-lombrozo/aidy/ai"
    "github.com/volodya-lombrozo/aidy/git"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "os"
)

func main() {
    // Read API key from configuration file
    homeDir, err := os.UserHomeDir()
    if err != nil {
        log.Fatalf("Error getting home directory: %v", err)
    }
    configPath := fmt.Sprintf("%s/.aidy.conf.yml", homeDir)
    configData, err := ioutil.ReadFile(configPath)
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    var config struct {
        OpenAIAPIKey string `yaml:"openai-api-key"`
    }

    err = yaml.Unmarshal(configData, &config)
    if err != nil {
        log.Fatalf("Error parsing config file: %v", err)
    }

    apiKey := config.OpenAIAPIKey
    if apiKey == "" {
        log.Fatalf("OpenAI API key not found in config file")
    }
    gitService := &git.RealGit{}

    branchName, err := gitService.GetBranchName()
    if err != nil {
        log.Fatalf("Error getting branch name: %v", err)
    }

    diff, err := gitService.GetDiff("main")
    if err != nil {
        log.Fatalf("Error getting git diff: %v", err)
    }
    fmt.Printf("Git Diff:\n%s\n", diff)
    // Use the OpenAI implementation
    aiService := ai.NewOpenAI(apiKey, "gpt-4o", 0.3)

    title, err := aiService.GenerateTitle(branchName, diff)
    if err != nil {
        log.Fatalf("Error generating title: %v", err)
    }

    body, err := aiService.GenerateBody(branchName, diff)
    if err != nil {
        log.Fatalf("Error generating body: %v", err)
    }

    fmt.Printf("Generated PR Command:\ngh pr create --title \"%s\" --body \"%s\"\n", title, body)
}
