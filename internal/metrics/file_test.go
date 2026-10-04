package metrics

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_ReturnsNoRunsWhenFileIsMissing(t *testing.T) {
	metrics := NewFile(filepath.Join(t.TempDir(), "metrics.json"))

	runs, err := metrics.Runs()

	require.NoError(t, err)
	assert.Empty(t, runs, "no runs expected for a missing file")
}

func TestFile_RecordsRunsInMissingFolder(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy", "metrics.json")
	metrics := NewFile(path)
	first := Run{Command: "commit", Time: time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC), Success: true, Duration: 1200}
	second := Run{Command: "start", Time: time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC), Success: false, Duration: 300}

	require.NoError(t, metrics.Record(first))
	require.NoError(t, metrics.Record(second))

	runs, err := NewFile(path).Runs()
	require.NoError(t, err)
	assert.Equal(t, []Run{first, second}, runs, "recorded runs should be read back in order")
}

func TestFile_FailsOnCorruptedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metrics.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0600))

	err := NewFile(path).Record(Run{Command: "commit"})

	assert.ErrorContains(t, err, "can't parse metrics")
}

func TestFile_FailsWhenPathIsFolder(t *testing.T) {
	path := t.TempDir()

	_, err := NewFile(path).Runs()

	assert.ErrorContains(t, err, "can't read metrics")
}
