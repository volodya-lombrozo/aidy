package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"

	"github.com/volodya-lombrozo/aidy/cache"
	"github.com/volodya-lombrozo/aidy/git"
)

type RealGithub struct {
	client     *http.Client
	baseURL    string
	gitService git.Git
	authToken  string
	ch         cache.AidyCache
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

func NewRealGithub(baseURL string, gitService git.Git, authToken string, ch cache.AidyCache) *RealGithub {
	return &RealGithub{
		client:     &http.Client{},
		baseURL:    baseURL,
		gitService: gitService,
		authToken:  authToken,
		ch:         ch,
	}
}

func (r *RealGithub) IssueDescription(number string) string {
	if r.gitService == nil {
		panic("Git service isn't set")
	}
	var task issue
	target := r.ch.Remote()
	if target != "" {
		url := fmt.Sprintf("%s/repos/%s/issues/%s", r.baseURL, target, number)
		log.Printf("Trying to get an issue description using the following URL: %s\n", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Sprintf("Error creating request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+r.authToken)
		resp, err := r.client.Do(req)
		if err != nil {
			return fmt.Sprintf("Error fetching issue description: %v", err)
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
			return fmt.Sprintf("Error reading response body: %v", err)
		}
		err = json.Unmarshal(body, &task)
		if err != nil {
			return fmt.Sprintf("Error unmarshaling issue JSON: %v", err)
		}
	} else {
		log.Fatalf("Cannot find a target repository to search for issue '%s'", number)
	}
	fmt.Printf("Title: %s\nBody: %s\n", task.Title, task.Body)
	return fmt.Sprintf("Title: '%s'\nBody: '%s'", task.Title, task.Body)
}

func (r *RealGithub) Labels() []string {
	if r.gitService == nil {
		panic("Git service isn't set")
	}
	var labels []label
	target := r.ch.Remote()
	if target != "" {
		log.Printf("Choosing labels from '%s'\n", target)
		url := fmt.Sprintf("%s/repos/%s/labels", r.baseURL, target)
		log.Printf("Trying to get repository labels using the following URL: %s\n", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Cannot create a new GET request to retrieve labels, because of '%v'", err)
		}
		req.Header.Set("Authorization", "Bearer "+r.authToken)
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

func (r *RealGithub) Remotes() []string {
	lines, err := r.gitService.Remotes()
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
