package ai

import (
    "context"
    "fmt"
    openai "github.com/sashabaranov/go-openai"
)

type OpenAI struct {
    client     *openai.Client
    model      string
    temperature float32
}

func NewOpenAI(apiKey, model string, temperature float32) *OpenAI {
    client := openai.NewClient(apiKey)
    return &OpenAI{
        client:     client,
        model:      model,
        temperature: temperature,
    }
}

func (o *OpenAI) GenerateTitle(branchName string) (string, error) {
    prompt := fmt.Sprintf(GenerateTitlePrompt, branchName)
    return o.generateText(prompt)
}

func (o *OpenAI) GenerateBody(branchName string) (string, error) {
    prompt := fmt.Sprintf(GenerateBodyPrompt, branchName)
    return o.generateText(prompt)
}

func (o *OpenAI) generateText(prompt string) (string, error) {
    req := openai.CompletionRequest{
        Model:       o.model,
        Prompt:      prompt,
        MaxTokens:   100,
        Temperature: o.temperature,
    }
    resp, err := o.client.CreateCompletion(context.Background(), req)
    if err != nil {
        return "", err
    }
    if len(resp.Choices) > 0 {
        return resp.Choices[0].Text, nil
    }
    return "", fmt.Errorf("no text generated")
}
