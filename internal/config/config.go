package config

type Config interface {
	Provider() (string, error)
	Model() (string, error)
	Token() (string, error)
	GithubKey() (string, error)
	Metrics() (bool, error)
	Jira() (Jira, error)
}

type Jira struct {
	URL   string `yaml:"url"`
	Email string `yaml:"email,omitempty"`
	Token string `yaml:"token"`
}

func (j Jira) Configured() bool {
	return j.URL != "" && j.Token != ""
}
