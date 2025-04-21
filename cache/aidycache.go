package cache

import (
    "log"
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
        log.Fatalf("Can't save the project remote address, because '%v'", err)
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
        log.Fatalf("Can't save the project summary, because '%v'", err)
    }
    err = a.ch.Set("summary-hash", hash)
    if err != nil {
        log.Fatalf("Can't save the project summary hash, because '%v'", err)
    }
}

type mockAidyCache struct {
}

func NewMockAidyCache() AidyCache {
    return &mockAidyCache{}
}

func (a *mockAidyCache) Remote() string {
    return "mock/remote"
}

func (a *mockAidyCache) WithRemote(remote string) {
}

func (a *mockAidyCache) Summary() (string, string) {
    return "mock summary", "mock hash"
}

func (a *mockAidyCache) WithSummary(summary, hash string) {
}
