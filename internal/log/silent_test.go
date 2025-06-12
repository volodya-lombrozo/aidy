package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSilent_Info(t *testing.T) {
	logger := NewSilent()
	logger.Info("This is an info message")
	// No assertion needed, as Silent does not log anything
	assert.NotNil(t, logger, "Expected logger to be created without error")
}
