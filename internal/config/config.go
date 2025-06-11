package config

type Config interface {
	Provider() (string, error)
	Model() (string, error)
	Token() (string, error)
	GithubKey() (string, error)
}
