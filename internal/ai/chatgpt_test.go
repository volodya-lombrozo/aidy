package ai

import (
	"context"
	"encoding/json"
	"errors"
	openai "github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChatGPTStream(t *testing.T) {
	for _, tc := range []struct {
		name, stream, want string
		failed             bool
	}{
		{"completed", "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Hello\"}\n\ndata: {\"type\":\"response.completed\"}\n\n", "Hello", false},
		{"partial failure", "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Partial\"}\n\ndata: {\"type\":\"response.failed\"}\n\n", "", true},
		{"disconnected", "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Partial\"}\n\n", "", true},
		{"incomplete", "data: {\"type\":\"response.incomplete\"}\n\n", "", true},
		{"empty", "data: {\"type\":\"response.completed\"}\n\n", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := readChatGPTStream(strings.NewReader(tc.stream))
			if tc.failed {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.want, out)
		})
	}
}

type chatGPTTransport func(*http.Request) (*http.Response, error)

func (f chatGPTTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestChatGPTRequests(t *testing.T) {
	for _, mode := range []string{"success", "access-error", "network-error", "unauthorized", "failed-stream", "bad-url"} {
		t.Run(mode, func(t *testing.T) {
			client := &chatGPTClient{endpoint: "https://example.com/responses", access: func(context.Context) (string, error) {
				if mode == "access-error" {
					return "", errors.New("sign in needed")
				}
				return "token", nil
			}}
			if mode == "bad-url" {
				client.endpoint = ":"
			}
			client.http = &http.Client{Transport: chatGPTTransport(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, "Bearer token", r.Header.Get("Authorization"))
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, false, body["store"])
				require.Equal(t, true, body["stream"])
				require.Equal(t, "prompt", body["instructions"])
				require.NotContains(t, body, "temperature")
				if mode == "network-error" {
					return nil, errors.New("offline")
				}
				status := 200
				if mode == "unauthorized" {
					status = 401
				}
				stream := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"result\"}\n\ndata: {\"type\":\"response.completed\"}\n\n"
				if mode == "failed-stream" {
					stream = "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":\"subscription_sharing_usage_limit_exceeded\"}}}\n\n"
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(stream)), Header: make(http.Header)}, nil
			})}
			response, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{Model: "model", Messages: []openai.ChatCompletionMessage{{Content: "prompt"}}})
			if mode == "success" {
				require.NoError(t, err)
				require.Equal(t, "result", response.Choices[0].Message.Content)
			} else {
				require.Error(t, err)
			}
		})
	}
}
