package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"github.com/volodya-lombrozo/aidy/cache"
	"github.com/volodya-lombrozo/aidy/git"
)

type github struct {
	client *http.Client
	url    string
	gs     git.Git
	token  string
	ch     cache.AidyCache
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

func NewGithub(baseURL string, gs git.Git, authToken string, ch cache.AidyCache) *github {
	return &github{
		client: &http.Client{},
		url:    baseURL,
		gs:     gs,
		token:  authToken,
		ch:     ch,
	}
}

func (r *github) Description(number string) string {
	if _, err := strconv.Atoi(number); err != nil {
		return fmt.Sprintf("Invalid issue number: '%s'", number)
	}
	var task issue
	target := r.ch.Remote()
	if target != "" {
		task = r.description(number, target)
	} else {
		log.Fatalf("Cannot find a target repository to search for issue '%s'", number)
	}
	fmt.Printf("Title: %s\nBody: %s\n", task.Title, task.Body)
	return fmt.Sprintf("Title: '%s'\nBody: '%s'", task.Title, task.Body)
}

func (r *github) Labels() []string {
	var labels []label
	target := r.ch.Remote()
	if target != "" {
		log.Printf("Choosing labels from '%s'\n", target)
		url := fmt.Sprintf("%s/repos/%s/labels", r.url, target)
		log.Printf("Trying to get repository labels using the following URL: %s\n", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Cannot create a new GET request to retrieve labels, because of '%v'", err)
		}
		req.Header.Set("Authorization", "Bearer "+r.token)
		resp, err := r.client.Do(req)
		if err != nil {
			log.Fatalf("Error fetching issue description: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("Error closing response body: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Skipping %s: Status %s\n", url, resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
		}
		err = json.Unmarshal(body, &labels)
		if err != nil {
			log.Fatalf("Error unmarshaling issue JSON: %v", err)
		}
	} else {
		log.Fatalf("Cannot determine where to get labels. Please set the target repository.")
	}
	var res []string
	for _, label := range labels {
		res = append(res, label.Name)
	}
	return res
}

func (r *github) Remotes() []string {
	lines, err := r.gs.Remotes()
	if err != nil {
		log.Fatalf("Cannot retrive git remotes: %v", err)
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
	return repos
}

// Here we retireve an issue or a PR description by thier number.
// GitHub uses the same URL structure for both issues and pull requests in the context of their API
// because every pull request is an issue under the hood.
// In other words, GET "/issues/:number" works for both issues and PRs.
func (r *github) description(number string, target string) issue {
	url := fmt.Sprintf("%s/repos/%s/issues/%s", r.url, target, number)
	log.Printf("Trying to get an issue description using the following URL: %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	resp, err := r.client.Do(req)
	if err != nil {
		log.Fatalf("Error fetching issue description: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Cannot retrieve issue using the following URL: '%s'. Response: '%s'", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	var task issue
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Fatalf("Error unmarshaling issue JSON: %v", err)
	}
	return task
}
