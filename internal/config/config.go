package config

type Config interface {
	OpenAiKey() (string, error)
	GithubKey() (string, error)
	DeepseekKey() (string, error)
	Model() (string, error)
}
