package ai

import (
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
