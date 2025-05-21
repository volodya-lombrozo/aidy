package config

type Config interface {
	GetOpenAIAPIKey() (string, error)
	GetGithubAPIKey() (string, error)
	GetDeepseekAPIKey() (string, error)
	GetModel() (string, error)
}
