package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/volodya-lombrozo/aidy/internal/chatgpt"
)

// NewChatGPT reuses aidy's prompts with a direct ChatGPT plan transport.
func NewChatGPT(model string, summary bool, language string) AI {
	if model == "" {
		model = "gpt-6.1-sol"
	}
	return NewOpenAIWithClient(&chatGPTClient{http: &http.Client{Timeout: 5 * time.Minute}, endpoint: "https://api.openai.com/v1/responses", access: chatGPTAccess}, model, 0, summary, language)
}
func chatGPTAccess(ctx context.Context) (string, error) {
	c, err := chatgpt.Default()
	if err != nil {
		return "", err
	}
	token, err := c.Access(ctx)
	if err == nil {
		return token, nil
	}
	// Never open a browser from scripts or CI.
	info, statErr := os.Stdin.Stat()
	if statErr != nil || info.Mode()&os.ModeCharDevice == 0 {
		return "", err
	}
	if loginErr := c.Login(ctx, os.Stderr); loginErr != nil {
		return "", loginErr
	}
	return c.Access(ctx)
}

type chatGPTClient struct {
	http     *http.Client
	endpoint string
	access   func(context.Context) (string, error)
}

func (c *chatGPTClient) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	var result openai.ChatCompletionResponse
	token, err := c.access(ctx)
	if err != nil {
		return result, err
	}
	var prompts []string
	for _, m := range request.Messages {
		prompts = append(prompts, m.Content)
	}
	body, err := json.Marshal(map[string]any{"model": request.Model, "instructions": strings.Join(prompts, "\n\n"), "input": []map[string]string{{"role": "user", "content": "Complete the requested text generation task."}}, "store": false, "stream": true})
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.http.Do(req)
	if err != nil {
		return result, errors.New("ChatGPT request failed; check your connection")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return result, fmt.Errorf("ChatGPT request failed (HTTP %d); check account access and limits at https://chatgpt.com/settings/usage", resp.StatusCode)
	}
	text, err := readChatGPTStream(resp.Body)
	if err != nil {
		return result, err
	}
	result.Choices = []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Role: "assistant", Content: text}}}
	return result, nil
}

func readChatGPTStream(in io.Reader) (string, error) {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	var output strings.Builder
	var data []string
	consume := func() (bool, error) {
		if len(data) == 0 {
			return false, nil
		}
		raw := strings.Join(data, "\n")
		data = nil
		if raw == "[DONE]" {
			return false, errors.New("ChatGPT stream ended without response.completed")
		}
		var event struct {
			Type, Delta string
			Response    struct{ Error *struct{ Code string } }
		}
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			return false, errors.New("invalid ChatGPT stream event")
		}
		switch event.Type {
		case "response.output_text.delta":
			output.WriteString(event.Delta)
		case "response.completed":
			return true, nil
		case "response.failed", "response.incomplete", "error":
			if event.Response.Error != nil && event.Response.Error.Code != "" {
				return false, fmt.Errorf("ChatGPT generation failed (%s); check ChatGPT plan access and usage limits", event.Response.Error.Code)
			}
			return false, fmt.Errorf("ChatGPT generation failed (%s); check ChatGPT plan access and usage limits", event.Type)
		}
		return false, nil
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			done, err := consume()
			if err != nil {
				return "", err
			}
			if done {
				if strings.TrimSpace(output.String()) == "" {
					return "", errors.New("no text generated")
				}
				return output.String(), nil
			}
		} else if strings.HasPrefix(line, "data:") {
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ChatGPT stream interrupted: %w", err)
	}
	if len(data) > 0 {
		done, err := consume()
		if err != nil {
			return "", err
		}
		if done && strings.TrimSpace(output.String()) != "" {
			return output.String(), nil
		}
	}
	return "", errors.New("ChatGPT stream ended without response.completed")
}
