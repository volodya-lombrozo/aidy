package cache

type mockCache struct {
}

func NewMockCache() Cache {
	return &mockCache{}
}

func (c *mockCache) Get(key string) (string, bool) {
	return "", true
}

func (c *mockCache) Set(key, value string) error {
	return nil
}
