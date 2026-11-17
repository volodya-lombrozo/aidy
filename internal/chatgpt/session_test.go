package chatgpt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRefreshPersistsRotation(t *testing.T) {
	c := &Client{Dir: t.TempDir()}
	c.HTTP = &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "registered", r.Form.Get("client_id"))
		require.Equal(t, "old-refresh", r.Form.Get("refresh_token"))
		require.Equal(t, resource, r.Form.Get("resource"))
		require.Empty(t, r.Form.Get("scope"))
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"access_token":"new-access","refresh_token":"new-refresh","expires_in":3600,"token_type":"Bearer"}`)), Header: make(http.Header)}, nil
	})}
	require.NoError(t, c.save(session{ClientID: "registered", Access: "old", Refresh: "old-refresh", Scope: permission, Expires: time.Now().Add(-time.Hour)}))
	token, err := c.Access(context.Background())
	require.NoError(t, err)
	require.Equal(t, "new-access", token)
	saved, err := c.read()
	require.NoError(t, err)
	require.Equal(t, "new-refresh", saved.Refresh)
	info, err := os.Stat(filepath.Join(c.Dir, "credentials.json"))
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}
}
func TestMissingSession(t *testing.T) {
	c := &Client{Dir: t.TempDir()}
	_, err := c.Access(context.Background())
	require.Error(t, err)
}

func TestIdentityRejectsUntrustedTokens(t *testing.T) {
	c := &Client{Dir: t.TempDir(), HTTP: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"keys":[]}`)), Header: make(http.Header)}, nil
	})}}
	for _, token := range []string{"", "eyJhbGciOiJub25lIn0.e30.", "eyJhbGciOiJSUzI1NiIsImtpZCI6Im1pc3NpbmcifQ.e30.YQ"} {
		_, _, err := c.identity(context.Background(), token, "registered", "nonce")
		require.Error(t, err)
	}
}

func TestLockIgnoresLegacyDirectory(t *testing.T) {
	c := &Client{Dir: t.TempDir()}
	require.NoError(t, os.Mkdir(filepath.Join(c.Dir, "lock"), 0700))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock, err := c.lock(ctx)
	require.NoError(t, err)
	unlock()
}

func TestLockWaitCanBeCancelled(t *testing.T) {
	c := &Client{Dir: t.TempDir()}
	unlock, err := c.lock(context.Background())
	require.NoError(t, err)
	defer unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = c.lock(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestLockReleasedAfterKilledProcess(t *testing.T) {
	c := &Client{Dir: t.TempDir()}
	child := exec.Command(os.Args[0], "-test.run=^TestLockChild$")
	child.Env = append(os.Environ(), "AIDY_TEST_LOCK_DIR="+c.Dir)
	stdout, err := child.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, child.Start())
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	ready := make(chan string, 1)
	go func() { scanner := bufio.NewScanner(stdout); scanner.Scan(); ready <- scanner.Text() }()
	select {
	case line := <-ready:
		require.Equal(t, "locked", line)
	case <-time.After(5 * time.Second):
		t.Fatal("child did not acquire the lock")
	}
	require.NoError(t, child.Process.Kill())
	_ = child.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock, err := c.lock(ctx)
	require.NoError(t, err)
	unlock()
}

func TestLockChild(t *testing.T) {
	dir := os.Getenv("AIDY_TEST_LOCK_DIR")
	if dir == "" {
		return
	}
	c := &Client{Dir: dir}
	unlock, err := c.lock(context.Background())
	require.NoError(t, err)
	defer unlock()
	fmt.Println("locked")
	time.Sleep(time.Minute)
}
