package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisabled_DoesNotRecord(t *testing.T) {
	origin := NewMock()
	metrics := NewDisabled(origin)

	require.NoError(t, metrics.Record(Run{Command: "commit"}))

	runs, err := origin.Runs()
	require.NoError(t, err)
	assert.Empty(t, runs, "disabled metrics should not record runs")
}

func TestDisabled_ReadsPreviousRuns(t *testing.T) {
	previous := Run{Command: "commit", Success: true}

	runs, err := NewDisabled(NewMock(previous)).Runs()

	require.NoError(t, err)
	assert.Equal(t, []Run{previous}, runs, "disabled metrics should still read recorded runs")
}
