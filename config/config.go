package config

type Config interface {
	GetOpenAIAPIKey() (string, error)
}
