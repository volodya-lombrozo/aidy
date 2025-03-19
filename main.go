package main

import (
    "fmt"
    "log"
    "github.com/volodya-lombrozo/aidy/ai"
    "github.com/volodya-lombrozo/aidy/git"
    "gopkg.in/yaml.v2"
    "io/ioutil"
)

func main() {
    // Read API key from configuration file
    configData, err := ioutil.ReadFile(".aidy.conf.yml")
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
    aiService := ai.NewOpenAI(apiKey, "text-davinci-003", 0.7)

    title, err := aiService.GenerateTitle(branchName)
    if err != nil {
        log.Fatalf("Error generating title: %v", err)
    }

    body, err := aiService.GenerateBody(branchName)
    if err != nil {
        log.Fatalf("Error generating body: %v", err)
    }

    fmt.Printf("Generated PR Command:\ngh pr create --title \"%s\" --body \"%s\"\n", title, body)
}
