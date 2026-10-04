package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRun_MarksSuccessfulRun(t *testing.T) {
	start := time.Now()

	run := NewRun("commit", start, nil)

	assert.Equal(t, "commit", run.Command)
	assert.Equal(t, start, run.Time)
	assert.True(t, run.Success, "run without error should be successful")
	assert.GreaterOrEqual(t, run.Duration, int64(0))
}

func TestNewRun_MarksFailedRun(t *testing.T) {
	run := NewRun("start", time.Now(), errors.New("boom"))

	assert.False(t, run.Success, "run with error should be failed")
}
