package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mapCache struct {
	store map[string]string
}

func (m *mapCache) Get(key string) (string, bool) {
	val, ok := m.store[key]
	return val, ok
}

func (m *mapCache) Set(key, value string) error {
	m.store[key] = value
	return nil
}

func TestAidyCache_Remote(t *testing.T) {
	mc := &mapCache{store: make(map[string]string)}
	ac := NewAidyCache(mc)

	err := mc.Set("target", "https://example.com/repo.git")
	require.NoError(t, err)
	remote := ac.Remote()
	assert.Equal(t, "https://example.com/repo.git", remote, "expected remote to be 'https://example.com/repo.git'")

	err = mc.Set("target", "")
	require.NoError(t, err)
	remote = ac.Remote()
	assert.Equal(t, "", remote, "expected remote to be ''")
}

func TestAidyCache_WithRemote(t *testing.T) {
	mc := &mapCache{store: make(map[string]string)}
	ac := NewAidyCache(mc)

	ac.WithRemote("https://example.com/repo.git")
	remote, ok := mc.Get("target")
	assert.True(t, ok)
	assert.Equal(t, "https://example.com/repo.git", remote, "expected remote to be 'https://example.com/repo.git'")
}

func TestAidyCache_Summary(t *testing.T) {
	mc := &mapCache{store: make(map[string]string)}
	ac := NewAidyCache(mc)

	err := mc.Set("summary", "Project Summary")
	require.NoError(t, err)
	err = mc.Set("summary-hash", "Project Summary Hash")
	require.NoError(t, err)
	summary, hash := ac.Summary()
    assert.Equal(t, "Project Summary", summary, "expected summary to be 'Project Summary'")
	assert.Equal(t, "Project Summary Hash", hash, "expected summary hash to be 'Project Summary Hash'")

	err = mc.Set("summary", "")
	require.NoError(t, err)
	err = mc.Set("summary-hash", "")
	require.NoError(t, err)
	summary, hash = ac.Summary()
	assert.Equal(t, "", summary, "expected summary to be ''")
	assert.Equal(t, "", hash, "expected summary to be ''")
}

func TestAidyCache_WithSummary(t *testing.T) {
	mc := &mapCache{store: make(map[string]string)}
	ac := NewAidyCache(mc)

	ac.WithSummary("Project Summary", "Project Summary Hash")
	summary, _ := mc.Get("summary")
	assert.Equal(t, "Project Summary", summary, "expected summary to be 'Project Summary'")
	hash, _ := mc.Get("summary-hash")
	assert.Equal(t, "Project Summary Hash", hash, "expected summary hash to be 'Project Summary Hash'")
}
