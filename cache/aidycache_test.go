package cache

import (
	"fmt"
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

type errorCache struct {
}

func (e *errorCache) Get(key string) (string, bool) {
	return "", false
}

func (e *errorCache) Set(key, value string) error {
	return fmt.Errorf("simulated error")
}

func TestAidyCache_WithRemote_Error(t *testing.T) {
	ac := NewAidyCache(&errorCache{})

	assert.Panics(t, func() {
		ac.WithRemote("https://example.com/repo.git")
	}, "expected panic due to error in Set")
}

func TestAidyCache_Summary_Error(t *testing.T) {
	ac := NewAidyCache(&errorCache{})

	summary, hash := ac.Summary()

	assert.Equal(t, "", summary, "expected summary to be '' due to error in Get")
	assert.Equal(t, "", hash, "expected hash to be '' due to error in Get")
}

func TestAidyCache_WithSummary_Error(t *testing.T) {
	ac := NewAidyCache(&errorCache{})

	assert.Panics(t, func() {
		ac.WithSummary("Project Summary", "Project Summary Hash")
	}, "expected panic due to error in Set")
}

func TestMockAidyCache_Remote(t *testing.T) {
	mc := NewMockAidyCache()
	remote := mc.Remote()
	assert.Equal(t, "mock/remote", remote, "expected remote to be 'mock/remote'")
}

func TestMockAidyCache_WithRemote(t *testing.T) {
	mc := NewMockAidyCache()
	mc.WithRemote("new/remote") // Should not change anything
	remote := mc.Remote()
	assert.Equal(t, "mock/remote", remote, "expected remote to be 'mock/remote'")
}

func TestMockAidyCache_Summary(t *testing.T) {
	mc := NewMockAidyCache()
	summary, hash := mc.Summary()
	assert.Equal(t, "mock summary", summary, "expected summary to be 'mock summary'")
	assert.Equal(t, "mock hash", hash, "expected hash to be 'mock hash'")
}

func TestMockAidyCache_WithSummary(t *testing.T) {
	mc := NewMockAidyCache()
	mc.WithSummary("new summary", "new hash") // Should not change anything
	summary, hash := mc.Summary()
	assert.Equal(t, "mock summary", summary, "expected summary to be 'mock summary'")
	assert.Equal(t, "mock hash", hash, "expected hash to be 'mock hash'")
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
