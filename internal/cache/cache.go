package cache

type Cache interface {
	Get(key string) (string, bool)
	Set(key, value string) error
}
