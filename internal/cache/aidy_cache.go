package cache

import (
	"fmt"
)

type AidyCache interface {
	Remote() string
	WithRemote(string)
	Summary() (string, string)
	WithSummary(string, string)
}

type aidyCache struct {
	ch Cache
}

func NewAidyCache(ch Cache) AidyCache {
	return &aidyCache{ch: ch}
}

func (a *aidyCache) Remote() string {
	repo, ok := a.ch.Get("target")
	if ok && repo != "" {
		return repo
	} else {
		return ""
	}
}

func (a *aidyCache) WithRemote(remote string) {
	err := a.ch.Set("target", remote)
	if err != nil {
		panic(fmt.Errorf("can't save the project remote address, because '%v'", err))
	}
}

func (a *aidyCache) Summary() (string, string) {
	summary, ok := a.ch.Get("summary")
	if !ok {
		return "", ""
	}
	hash, ok := a.ch.Get("summary-hash")
	if !ok {
		return "", ""
	}
	return summary, hash
}

func (a *aidyCache) WithSummary(summary string, hash string) {
	err := a.ch.Set("summary", summary)
	if err != nil {
		panic(fmt.Errorf("can't save the project summary, because '%v'", err))
	}
	err = a.ch.Set("summary-hash", hash)
	if err != nil {
		panic(fmt.Errorf("can't save the project summary hash, because '%v'", err))
	}
}

type mockAidyCache struct {
	inner map[string]string
}

func NewMockAidyCache() AidyCache {
	memory := make(map[string]string)
	memory["target"] = "mock/remote"
	memory["summary"] = "mock summary"
	memory["summary-hash"] = "mock hash"
	return &mockAidyCache{inner: memory}
}

func (a *mockAidyCache) Remote() string {
	return a.inner["target"]
}

func (a *mockAidyCache) WithRemote(remote string) {
	a.inner["target"] = remote
}

func (a *mockAidyCache) Summary() (string, string) {
	return a.inner["summary"], a.inner["summary-hash"]
}

func (a *mockAidyCache) WithSummary(summary, hash string) {
	a.inner["summary"] = summary
	a.inner["summary-hash"] = hash
}
