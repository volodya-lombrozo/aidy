package chatgpt

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Login opens the browser and stores credentials only after validating identity.
func (c *Client) Login(ctx context.Context, out io.Writer) error {
	unlock, err := c.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	s, err := c.read()
	if err != nil {
		return err
	}
	if s.Host == "" {
		b := random()[:32]
		s.Host = fmt.Sprintf("urn:uuid:%s-%s-4%s-8%s-%s", b[:8], b[8:12], b[13:16], b[17:20], b[20:])
		if err = c.save(s); err != nil {
			return err
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer func() { _ = listener.Close() }()
	redirect := "http://" + listener.Addr().String() + "/auth/callback"
	state, nonce, verifier := random(), random(), random()
	hash := sha256.Sum256([]byte(verifier))
	clientID := s.ClientID
	if clientID == "" {
		clientID = "dynamic_agent_client"
	}
	q := url.Values{"client_id": {clientID}, "ext_agent_host_id": {s.Host}, "response_type": {"code"}, "redirect_uri": {redirect}, "scope": {"openid profile email offline_access resource.invoke chatgpt.tokens.use.direct"}, "resource": {resource}, "state": {state}, "nonce": {nonce}, "code_challenge_method": {"S256"}, "code_challenge": {base64.RawURLEncoding.EncodeToString(hash[:])}}
	if s.ClientID == "" {
		q.Set("agent_name_hint", "aidy")
	}
	// Do not include ID-token hints in URLs printed to the terminal.
	results := make(chan url.Values, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid sign-in callback", 400)
			return
		}
		select {
		case results <- r.URL.Query():
			_, _ = fmt.Fprintln(w, "Sign-in received. Return to aidy to finish.")
		default:
			http.Error(w, "Callback already received", 400)
		}
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	defer func() { _ = server.Close() }()
	go func() { _ = server.Serve(listener) }()
	address := issuer + "/api/accounts/authorize?" + q.Encode()
	if _, err := fmt.Fprintf(out, "Continue with ChatGPT in your browser:\n%s\n", address); err != nil {
		return err
	}
	open := c.browser
	if open == nil {
		open = openBrowser
	}
	_ = open(ctx, address) // The printed URL also supports manual browser opening.

	var result url.Values
	select {
	case result = <-results:
	case <-ctx.Done():
		return ctx.Err()
	}
	if result.Get("error") != "" {
		return errors.New("ChatGPT authorization declined or failed")
	}
	issued := result.Get("client_id")
	if s.ClientID != "" {
		if issued != "" && issued != s.ClientID {
			return errors.New("ChatGPT callback changed the registered client")
		}
		issued = s.ClientID
	}
	if issued == "" || issued == "dynamic_agent_client" || result.Get("code") == "" {
		return errors.New("incomplete ChatGPT registration callback")
	}
	t, err := c.exchange(ctx, url.Values{"grant_type": {"authorization_code"}, "client_id": {issued}, "code": {result.Get("code")}, "code_verifier": {verifier}, "redirect_uri": {redirect}})
	if err != nil {
		return err
	}
	subject, email, err := c.identity(ctx, t.ID, issued, nonce)
	if err != nil {
		return err
	}
	if s.Subject != "" && s.Subject != subject {
		return errors.New("ChatGPT identity differs from the registered account")
	}
	if !allowed(t.Scope) {
		return errors.New("ChatGPT plan permission was not granted")
	}
	s.ClientID = issued
	s.Subject = subject
	s.Email = email
	apply(&s, t)
	if err = c.save(s); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "ChatGPT connected: %s\n", email)
	return err
}

func openBrowser(ctx context.Context, address string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", address)
	case "windows":
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", address)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", address)
	}
	return cmd.Run()
}

// identity verifies RS256 against the issuer's published keys before using claims.
func (c *Client) identity(ctx context.Context, token, client, nonce string) (string, string, error) {
	fail := errors.New("ChatGPT ID token validation failed")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", fail
	}
	decode := func(part string, v any) error {
		b, e := base64.RawURLEncoding.DecodeString(part)
		if e != nil {
			return e
		}
		return json.Unmarshal(b, v)
	}
	var header struct{ Alg, Kid string }
	if decode(parts[0], &header) != nil || header.Alg != "RS256" {
		return "", "", fail
	}
	req, err := http.NewRequestWithContext(ctx, "GET", issuer+"/.well-known/jwks.json", nil)
	if err != nil {
		return "", "", err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", "", errors.New("could not retrieve ChatGPT identity keys")
	}
	defer func() { _ = resp.Body.Close() }()
	var keys struct {
		Keys []struct{ Kid, Kty, N, E string }
	}
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&keys) != nil {
		return "", "", fail
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", "", fail
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	valid := false
	for _, k := range keys.Keys {
		if k.Kid != header.Kid || k.Kty != "RSA" {
			continue
		}
		n, e1 := base64.RawURLEncoding.DecodeString(k.N)
		e, e2 := base64.RawURLEncoding.DecodeString(k.E)
		if e1 != nil || e2 != nil || len(e) > 4 {
			continue
		}
		exponent := 0
		for _, b := range e {
			exponent = exponent*256 + int(b)
		}
		if rsa.VerifyPKCS1v15(&rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exponent}, crypto.SHA256, digest[:], signature) == nil {
			valid = true
			break
		}
	}
	if !valid {
		return "", "", fail
	}
	var claims struct {
		Iss, Sub, Email, Nonce string
		Aud                    json.RawMessage
		Exp, Iat               int64
	}
	if decode(parts[1], &claims) != nil {
		return "", "", fail
	}
	var audience string
	var audiences []string
	if json.Unmarshal(claims.Aud, &audience) == nil {
		audiences = []string{audience}
	} else {
		_ = json.Unmarshal(claims.Aud, &audiences)
	}
	matched := false
	for _, a := range audiences {
		if a == client {
			matched = true
		}
	}
	if !matched || claims.Iss != issuer || claims.Sub == "" || claims.Nonce != nonce || claims.Exp <= time.Now().Unix() || claims.Iat <= 0 || claims.Iat > time.Now().Add(5*time.Second).Unix() {
		return "", "", fail
	}
	return claims.Sub, claims.Email, nil
}
