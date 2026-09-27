package output

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditor_FindEditor(t *testing.T) {
	original := os.Getenv("EDITOR")
	defer func() {
		if err := os.Setenv("EDITOR", original); err != nil {
			t.Errorf("failed to reset EDITOR environment variable: %v", err)
		}
	}()
	tests := []struct {
		envEditor string
		os        string
		expected  string
	}{
		{"", "windows", "notepad"},
		{"", "darwin", "vi"},
		{"", "linux", "vi"},
		{"nano", "linux", "nano"},
		{"code", "darwin", "code"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("EDITOR=%s OS=%s", test.envEditor, test.os), func(t *testing.T) {
			if err := os.Setenv("EDITOR", test.envEditor); err != nil {
				require.NoError(t, err, "failed to set EDITOR environment variable")
			}
			found := findEditor(test.os)
			assert.Equal(t, test.expected, found)
		})
	}
}
