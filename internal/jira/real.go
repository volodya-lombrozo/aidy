package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/log"
)

type jira struct {
	client *http.Client
	url    string
	email  string
	token  string
	log    log.Logger
}

type issue struct {
	Fields struct {
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
	} `json:"fields"`
}

func NewJira(url string, email string, token string) *jira {
	return &jira{
		client: &http.Client{},
		url:    strings.TrimSuffix(url, "/"),
		email:  email,
		token:  token,
		log:    log.Default(),
	}
}

func (j *jira) IssueType(key string) (string, error) {
	url := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=issuetype", j.url, key)
	j.log.Debug("retrieving jira issue type using the following url: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create a new GET request to retrieve jira issue '%s': %w", key, err)
	}
	j.authorize(req)
	resp, err := j.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching jira issue '%s': %w", key, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			j.log.Error("error closing response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cannot retrieve jira issue '%s' using the following url: '%s'. response: '%s'", key, url, resp.Status)
	}
	var task issue
	if err = json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return "", fmt.Errorf("error unmarshaling jira issue json: %w", err)
	}
	return task.Fields.IssueType.Name, nil
}

func (j *jira) authorize(req *http.Request) {
	if j.email == "" {
		req.Header.Set("Authorization", "Bearer "+j.token)
	} else {
		req.SetBasicAuth(j.email, j.token)
	}
}
