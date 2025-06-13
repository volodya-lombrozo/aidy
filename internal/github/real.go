package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"github.com/volodya-lombrozo/aidy/internal/cache"
	"github.com/volodya-lombrozo/aidy/internal/git"
	"github.com/volodya-lombrozo/aidy/internal/log"
)

type github struct {
	client *http.Client
	url    string
	git    git.Git
	token  string
	ch     cache.AidyCache
	log    log.Logger
}

type issue struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type label struct {
	Id          int64  `json:"id"`
	Node        string `json:"node_id"`
	Url         string `json:"url"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

func NewGithub(url string, gs git.Git, token string, ch cache.AidyCache) *github {
	return &github{
		client: &http.Client{},
		url:    url,
		git:    gs,
		token:  token,
		ch:     ch,
		log:    log.Get(),
	}
}

func (r *github) Description(number string) (string, error) {
	if _, err := strconv.Atoi(number); err != nil {
		return fmt.Sprintf("invalid issue number: '%s'", number), nil
	}
	var task issue
	target := r.ch.Remote()
	if target != "" {
		var err error
		task, err = r.description(number, target)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("cannot find a target repository to search for issue '%s'", number)
	}
	r.log.Debug("retrieved issue description for #%s", number)
	r.log.Debug("issue title: '%s'", task.Title)
	r.log.Debug("issue body: '%s'", task.Body)
	return fmt.Sprintf("Title: '%s'\nBody: '%s'", task.Title, task.Body), nil
}

func (r *github) Labels() ([]string, error) {
	var labels []label
	target := r.ch.Remote()
	if target != "" {
		r.log.Debug("retrieving labels for target repository '%s'", target)
		url := fmt.Sprintf("%s/repos/%s/labels", r.url, target)
		r.log.Debug("using the following URL to retrieve labels: %s", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create a new GET request to retrieve labels: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+r.token)
		resp, err := r.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching issue labels: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				r.log.Error("error closing response body: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNotFound {
				return []string{}, nil
			}
			return nil, fmt.Errorf("cannot retrieve labels using the following url: '%s'. response: '%s'", url, resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		err = json.Unmarshal(body, &labels)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling issue json: %w", err)
		}
	} else {
		return nil, fmt.Errorf("cannot determine where to get labels, please set the target repository")
	}
	var res []string
	for _, label := range labels {
		res = append(res, label.Name)
	}
	return res, nil
}

func (r *github) Remotes() ([]string, error) {
	lines, err := r.git.Remotes()
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve git remotes: %w", err)
	}
	re := regexp.MustCompile(`(?:git@github\.com:|https://github\.com/)([^/]+/[^.]+)(?:\.git)?`)
	unique := make(map[string]struct{})
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			unique[string(matches[1])] = struct{}{}
		}
	}
	var repos []string
	for repo := range unique {
		repos = append(repos, repo)
	}
	sort.Strings(repos)
	return repos, nil
}

// Here we retireve an issue or a PR description by thier number.
// GitHub uses the same URL structure for both issues and pull requests in the context of their API
// because every pull request is an issue under the hood.
// In other words, GET "/issues/:number" works for both issues and PRs.
func (r *github) description(number string, target string) (issue, error) {
	url := fmt.Sprintf("%s/repos/%s/issues/%s", r.url, target, number)
	r.log.Debug("trying to get an issue description using the following url: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return issue{}, fmt.Errorf("cannot create a new GET request to retrieve issue description: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	resp, err := r.client.Do(req)
	if err != nil {
		return issue{}, fmt.Errorf("error fetching issue description: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return issue{}, fmt.Errorf("cannot retrieve issue using the following url: '%s'. response: '%s'", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return issue{}, fmt.Errorf("error reading response body: %w", err)
	}
	var task issue
	err = json.Unmarshal(body, &task)
	if err != nil {
		return issue{}, fmt.Errorf("error unmarshaling issue json: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return issue{}, fmt.Errorf("error closing response body: %w", err)
	}
	return task, nil
}
