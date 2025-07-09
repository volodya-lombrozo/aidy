package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAndGetLogger(t *testing.T) {
	mock := NewMock()
	Set(mock)

	initialised := Default()

	assert.Equal(t, mock, initialised, "Expected retrieved logger to be the mock logger")
}

func TestSetLoggerNilPanics(t *testing.T) {
	assert.Panics(t, func() {
		Set(nil)
	}, "Expected panic when setting logger to nil")
}

func TestGetLoggerNotSetPanics(t *testing.T) {
	original := main
	defer func() {
		main = original
	}()
	main = nil
	require.Panics(t, func() {
		Default()
	}, "Expected panic when getting logger that is not set")
}
