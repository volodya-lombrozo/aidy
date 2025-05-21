package ai

// import (
//     "context"
//     "fmt"
//     "testing"
//     openai "github.com/sashabaranov/go-openai"
// )
//
// type MockOpenAIClient struct{}
//
// func (m *MockOpenAIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
//     if req.Messages[0].Content == "test prompt" {
//         return openai.ChatCompletionResponse{
//             Choices: []openai.ChatCompletionChoice{
//                 {Message: openai.ChatCompletionMessage{Content: "Mock response"}},
//             },
//         }, nil
//     }
//     return openai.ChatCompletionResponse{}, fmt.Errorf("mock error")
// }
//
// func TestGenerateTitle(t *testing.T) {
//     mockClient := &MockOpenAIClient{}
//     openAI := &MyOpenAI{
//         client:     mockClient,
//         model:      "test-model",
//         temperature: 0.5,
//     }
//
//     title, err := openAI.GenerateTitle("123_feature", "test diff")
//     if err != nil {
//         t.Fatalf("Expected no error, got %v", err)
//     }
//     if title != "Mock response" {
//         t.Fatalf("Expected 'Mock response', got '%s'", title)
//     }
// }
//
// func TestGenerateBody(t *testing.T) {
//     mockClient := &MockOpenAIClient{}
//     openAI := &MyOpenAI{
//         client:     mockClient,
//         model:      "test-model",
//         temperature: 0.5,
//     }
//
//     body, err := openAI.GenerateBody("123_feature", "test diff")
//     if err != nil {
//         t.Fatalf("Expected no error, got %v", err)
//     }
//     if body != "Mock response" {
//         t.Fatalf("Expected 'Mock response', got '%s'", body)
//     }
// }
