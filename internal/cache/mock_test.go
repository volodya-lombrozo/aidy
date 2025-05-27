package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockCache_ReturnsEmpty(t *testing.T) {
	ch := NewMockCache()

	value, ok := ch.Get("any")

	assert.True(t, ok, "expected ok to be true")
	assert.Equal(t, "", value, "expected value to be empty")
}
