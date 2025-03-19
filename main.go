package main

import (
    "fmt"
    "log"
    "github.com/volodya-lombrozo/aidy/ai"
    "github.com/volodya-lombrozo/aidy/git"
)

func main() {
    // Use the mock Git implementation
    gitService := &git.MockGit{}

    branchName, err := gitService.GetBranchName()
    if err != nil {
        log.Fatalf("Error getting branch name: %v", err)
    }

    // Use the mock AI implementation
    aiService := &ai.MockAI{}

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
