package ai

import (
    "context"
    "fmt"
    "strings"
    openai "github.com/sashabaranov/go-openai"
)

type OpenAI struct {
    client     *openai.Client
    model      string
    temperature float32
}

func extractIssueNumber(branchName string) string {
    // Assuming the branch name format is "<issue-number>_<description>"
    parts := strings.Split(branchName, "_")
    if len(parts) > 0 {
        return parts[0]
    }
    return "unknown"
}

func NewOpenAI(apiKey, model string, temperature float32) *OpenAI {
    client := openai.NewClient(apiKey)
    return &OpenAI{
        client:     client,
        model:      model,
        temperature: temperature,
    }
}

func (o *OpenAI) GenerateTitle(branchName, diff string) (string, error) {
    // Extract issue number from branch name
    issueNumber := extractIssueNumber(branchName)
    prompt := fmt.Sprintf(GenerateTitlePrompt, issueNumber, diff)
    return o.generateText(prompt)
}

func (o *OpenAI) GenerateBody(branchName, diff string) (string, error) {
    prompt := fmt.Sprintf(GenerateBodyPrompt, branchName, diff)
    return o.generateText(prompt)
}

func (o *OpenAI) generateText(prompt string) (string, error) {
    fmt.Println("Prompt being sent to OpenAI:")
    fmt.Println(prompt)

    req := openai.ChatCompletionRequest{
        Model: o.model,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    "system",
                Content: prompt,
            },
        },
        MaxTokens:   100,
        Temperature: o.temperature,
    }
    resp, err := o.client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        return "", err
    }
    if len(resp.Choices) > 0 {
        return resp.Choices[0].Message.Content, nil
    }
    return "", fmt.Errorf("no text generated")
}
