package aidy

import (
	"net/url"
	"regexp"
	"strings"
)

type repository struct {
	host string
	path string
	url  string
}

var scpRemote = regexp.MustCompile(`^[^@/]+@([^:/]+):(.+)$`)

func parseRepository(remote string) (repository, bool) {
	fields := strings.Fields(remote)
	if len(fields) > 1 {
		remote = fields[1]
	}
	if matches := scpRemote.FindStringSubmatch(remote); len(matches) == 3 {
		remote = "https://" + matches[1] + "/" + matches[2]
	}
	parsed, err := url.Parse(remote)
	if err != nil || parsed.Host == "" {
		return repository{}, false
	}
	switch parsed.Scheme {
	case "ssh":
		parsed.Scheme = "https"
		parsed.Host = parsed.Hostname()
	case "https", "http":
	default:
		return repository{}, false
	}
	parsed.Host = strings.ToLower(parsed.Host)
	path := strings.TrimSuffix(strings.Trim(parsed.Path, "/"), ".git")
	if !strings.Contains(path, "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return repository{}, false
	}
	canonical := (&url.URL{Scheme: parsed.Scheme, Host: parsed.Host, Path: "/" + path}).String()
	return repository{host: parsed.Host, path: path, url: canonical}, true
}
