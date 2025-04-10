package github

import (
	"encoding/json"
	"fmt"
	"github.com/volodya-lombrozo/aidy/git"
	"io"
	"log"
	"net/http"
	"strings"
)

type RealGithub struct {
	client     *http.Client
	baseURL    string
	gitService git.Git
	authToken  string
}

type Issue struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func NewRealGithub(baseURL string, gitService git.Git, authToken string) *RealGithub {
	return &RealGithub{
		client:     &http.Client{},
		baseURL:    baseURL,
		gitService: gitService,
		authToken:  authToken,
	}
}

func (r *RealGithub) IssueDescription(number string) string {
	if r.gitService == nil {
		panic("Git service isn't set")
	}
	urls, gerr := r.gitService.GetAllRemoteURLs()
	if gerr != nil {
		panic(gerr)
	}
	credentials := parseOwnerRepoPairs(urls)
    var issue Issue
    for _, cred := range credentials {
        url := fmt.Sprintf("%s/repos/%s/%s/issues/%s", r.baseURL, cred[0], cred[1], number)
        log.Printf("Tying to get an isse description by using the following url: %s\n", url)
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
            fmt.Printf("Skipping %s: status %s\n", url, resp.Status)
            continue
        }
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return fmt.Sprintf("Error reading response body: %v", err)
        }
        err = json.Unmarshal(body, &issue)
        if err != nil {
            return fmt.Sprintf("Error unmarshaling issue JSON: %v", err)
        }
    }
	fmt.Printf("Title: %s\nBody: %s\n", issue.Title, issue.Body)
	return fmt.Sprintf("Title: '%s'\nBody: '%s'", issue.Title, issue.Body)
}

func parseOwnerRepoPairs(urls []string) [][2]string {
	var pairs [][2]string
	seen := make(map[string]struct{})

	for _, url := range urls {
		var owner, repo string

		switch {
		case strings.HasPrefix(url, "git@"):
			// SSH format: git@github.com:owner/repo.git
			parts := strings.SplitN(url, ":", 2)
			if len(parts) != 2 {
				continue
			}
			pathParts := strings.Split(strings.TrimSuffix(parts[1], ".git"), "/")
			if len(pathParts) != 2 {
				continue
			}
			owner, repo = pathParts[0], pathParts[1]

		case strings.HasPrefix(url, "https://"):
			// HTTPS format: https://github.com/owner/repo.git
			p := strings.TrimPrefix(url, "https://github.com/")
			pathParts := strings.Split(strings.TrimSuffix(p, ".git"), "/")
			if len(pathParts) != 2 {
				continue
			}
			owner, repo = pathParts[0], pathParts[1]

		default:
			continue
		}

		key := owner + "/" + repo
		if _, ok := seen[key]; !ok {
			pairs = append(pairs, [2]string{owner, repo})
			seen[key] = struct{}{}
		}
	}

	return pairs
}
