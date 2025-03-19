package main

import (
    "fmt"
    "log"
    "yourmodule/ai"
)

func main() {
    branchName := "feature/123-add-new-feature" // Example branch name

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
