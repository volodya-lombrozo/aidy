// Package chatgpt manages locally stored ChatGPT plan credentials.
package chatgpt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const issuer = "https://auth.openai.com"
const resource = "https://api.openai.com/v1"
const permission = "chatgpt.tokens.use.direct"

// Client owns one local credential store. HTTP is injectable for tests.
type Client struct {
	Dir     string
	HTTP    *http.Client
	browser func(context.Context, string) error
}
type session struct {
	Host     string    `json:"host"`
	ClientID string    `json:"client_id"`
	Subject  string    `json:"subject"`
	Email    string    `json:"email"`
	Access   string    `json:"access_token"`
	Refresh  string    `json:"refresh_token"`
	ID       string    `json:"id_token"`
	Scope    string    `json:"scope"`
	Expires  time.Time `json:"expires_at"`
}
type tokens struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	ID      string `json:"id_token"`
	Scope   string `json:"scope"`
	Type    string `json:"token_type"`
	Expires int64  `json:"expires_in"`
}

func Default() (*Client, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &Client{Dir: filepath.Join(home, ".aidy", "chatgpt"), HTTP: &http.Client{Timeout: 5 * time.Minute}}, nil
}
func (c *Client) lock(ctx context.Context) (func(), error) {
	if err := os.MkdirAll(c.Dir, 0700); err != nil {
		return nil, err
	}
	// Kernel locks are released even when the process is interrupted or killed.
	// Use a different path from the legacy directory lock, which can be stale.
	file, err := os.OpenFile(filepath.Join(c.Dir, "credentials.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	for {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return nil, err
		}
		acquired, err := tryLock(file)
		if err != nil {
			_ = file.Close()
			return nil, err
		}
		if acquired {
			return func() { _ = file.Close() }, nil
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, fmt.Errorf("waiting for ChatGPT credential lock: %w", ctx.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
}
func (c *Client) read() (session, error) {
	var s session
	b, err := os.ReadFile(filepath.Join(c.Dir, "credentials.json"))
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(b, &s)
	return s, err
}
func (c *Client) save(s session) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(c.Dir, ".credentials-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err = f.Write(b); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(c.Dir, "credentials.json"))
}
func random() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func allowed(scope string) bool {
	for _, s := range strings.Fields(scope) {
		if s == permission {
			return true
		}
	}
	return false
}
func (c *Client) exchange(ctx context.Context, form url.Values) (tokens, error) {
	var t tokens
	form.Set("resource", resource)
	req, err := http.NewRequestWithContext(ctx, "POST", issuer+"/api/accounts/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return t, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return t, errors.New("ChatGPT token request failed; check your connection")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return t, fmt.Errorf("ChatGPT token request failed (HTTP %d); sign in by running an aidy AI command in an interactive terminal", resp.StatusCode)
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&t); err != nil {
		return t, errors.New("invalid ChatGPT token response")
	}
	if t.Access == "" || t.Expires <= 0 || !strings.EqualFold(t.Type, "Bearer") {
		return t, errors.New("incomplete ChatGPT token response")
	}
	return t, nil
}
func apply(s *session, t tokens) {
	s.Access = t.Access
	s.Expires = time.Now().Add(time.Duration(t.Expires) * time.Second)
	if t.Refresh != "" {
		s.Refresh = t.Refresh
	}
	if t.ID != "" {
		s.ID = t.ID
	}
	if t.Scope != "" {
		s.Scope = t.Scope
	}
}

// Access returns a current access token, refreshing under a process lock.
func (c *Client) Access(ctx context.Context) (string, error) {
	unlock, err := c.lock(ctx)
	if err != nil {
		return "", err
	}
	defer unlock()
	s, err := c.read()
	if err != nil {
		return "", err
	}
	if s.Access == "" {
		return "", errors.New("ChatGPT is not signed in; sign in by running an aidy AI command in an interactive terminal")
	}
	if !allowed(s.Scope) {
		return "", errors.New("ChatGPT plan permission missing; sign in by running an aidy AI command in an interactive terminal")
	}
	if time.Until(s.Expires) > time.Minute {
		return s.Access, nil
	}
	if s.Refresh == "" {
		return "", errors.New("ChatGPT session expired; sign in by running an aidy AI command in an interactive terminal")
	}
	t, err := c.exchange(ctx, url.Values{"grant_type": {"refresh_token"}, "client_id": {s.ClientID}, "refresh_token": {s.Refresh}})
	if err != nil {
		return "", err
	}
	apply(&s, t)
	if !allowed(s.Scope) {
		return "", errors.New("ChatGPT plan permission was revoked; sign in by running an aidy AI command in an interactive terminal")
	}
	if err = c.save(s); err != nil {
		return "", err
	}
	return s.Access, nil
}
