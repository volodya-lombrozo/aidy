package github

type Github interface {
	Description(number string) string
	Labels() []string
	Remotes() []string
}
