package chatgpt

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func signedIdentity(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()
	enc := func(v any) string {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	unsigned := enc(map[string]string{"alg": "RS256", "kid": "test"}) + "." + enc(claims)
	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	require.NoError(t, err)
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}
func keyResponse(t *testing.T, key *rsa.PrivateKey) *http.Response {
	t.Helper()
	b, err := json.Marshal(map[string]any{"keys": []map[string]string{{"kid": "test", "kty": "RSA", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())}}})
	require.NoError(t, err)
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(b))), Header: make(http.Header)}
}
func TestLogin(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	for _, mode := range []string{"new", "returning", "denied", "missing-client", "changed-client", "wrong-account", "missing-permission", "exchange-failure", "invalid-identity"} {
		t.Run(mode, func(t *testing.T) {
			c := &Client{Dir: t.TempDir()}
			var auth url.Values
			if mode == "returning" || mode == "changed-client" || mode == "wrong-account" {
				require.NoError(t, c.save(session{Host: "saved-host", ClientID: "registered", Subject: "user"}))
			}
			c.browser = func(ctx context.Context, address string) error {
				parsed, e := url.Parse(address)
				require.NoError(t, e)
				auth = parsed.Query()
				require.Equal(t, "S256", auth.Get("code_challenge_method"))
				require.NotEmpty(t, auth.Get("ext_agent_host_id"))
				callback, e := url.Parse(auth.Get("redirect_uri"))
				require.NoError(t, e)
				// A forged state must not end the pending sign-in.
				bad := *callback
				bad.RawQuery = url.Values{"state": {"forged"}}.Encode()
				resp, e := http.Get(bad.String())
				require.NoError(t, e)
				require.Equal(t, 400, resp.StatusCode)
				_ = resp.Body.Close()
				q := url.Values{"state": {auth.Get("state")}, "code": {"code"}, "client_id": {"registered"}}
				if mode == "denied" {
					q.Set("error", "access_denied")
				}
				if mode == "missing-client" {
					q.Del("client_id")
				}
				if mode == "changed-client" {
					q.Set("client_id", "other")
				}
				if mode == "returning" {
					q.Del("client_id")
					require.Empty(t, auth.Get("agent_name_hint"))
				}
				callback.RawQuery = q.Encode()
				resp, e = http.Get(callback.String())
				require.NoError(t, e)
				_ = resp.Body.Close()
				return nil
			}
			c.HTTP = &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				if strings.HasSuffix(r.URL.Path, "jwks.json") {
					return keyResponse(t, key), nil
				}
				require.NoError(t, r.ParseForm())
				require.Equal(t, "registered", r.Form.Get("client_id"))
				require.Equal(t, auth.Get("redirect_uri"), r.Form.Get("redirect_uri"))
				digest := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
				require.Equal(t, auth.Get("code_challenge"), base64.RawURLEncoding.EncodeToString(digest[:]))
				if mode == "exchange-failure" {
					return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(""))}, nil
				}
				sub := "user"
				if mode == "wrong-account" {
					sub = "other"
				}
				claims := map[string]any{"iss": issuer, "sub": sub, "email": "user@example.com", "aud": []string{"registered"}, "nonce": auth.Get("nonce"), "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
				id := signedIdentity(t, key, claims)
				if mode == "invalid-identity" {
					id = "invalid"
				}
				scope := permission
				if mode == "missing-permission" {
					scope = "openid"
				}
				b, e := json.Marshal(map[string]any{"access_token": "access", "refresh_token": "refresh", "id_token": id, "scope": scope, "token_type": "Bearer", "expires_in": 3600})
				require.NoError(t, e)
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(b)))}, nil
			})}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err := c.Login(ctx, io.Discard)
			if mode == "new" || mode == "returning" {
				require.NoError(t, err)
				s, e := c.read()
				require.NoError(t, e)
				require.Equal(t, "user", s.Subject)
				require.Equal(t, "registered", s.ClientID)
				require.Equal(t, "access", s.Access)
			} else {
				require.Error(t, err)
				s, e := c.read()
				require.NoError(t, e)
				require.Empty(t, s.Access)
			}
		})
	}
}

func TestIdentityClaims(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	c := &Client{HTTP: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { return keyResponse(t, key), nil })}}
	for _, field := range []string{"valid", "iss", "sub", "nonce", "aud", "exp", "iat"} {
		t.Run(field, func(t *testing.T) {
			claims := map[string]any{"iss": issuer, "sub": "user", "email": "email", "aud": "registered", "nonce": "nonce", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
			if field == "exp" || field == "iat" {
				claims[field] = int64(0)
			} else if field != "valid" {
				claims[field] = ""
			}
			sub, _, e := c.identity(context.Background(), signedIdentity(t, key, claims), "registered", "nonce")
			if field == "valid" {
				require.NoError(t, e)
				require.Equal(t, "user", sub)
			} else {
				require.Error(t, e)
			}
		})
	}
}
