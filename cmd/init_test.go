package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/config"
)

func TestInit_Help(t *testing.T) {
	var out bytes.Buffer
	command := newInitCmd()
	command.SetOut(&out)
	command.SetArgs([]string{"--help"})

	err := command.Execute()

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "Create a ~/.aidy.conf.yml configuration file")
}

func TestRunInit_CreatesConfigFileByProviderNumber(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy.conf.yml")
	in := strings.NewReader("2\n\nsk-test-openai-key\n")
	var out bytes.Buffer

	err := runInit(in, &out, path)

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "configuration file created at")
	conf, err := config.YamlConf(path)
	require.NoError(t, err, "config file should be readable")
	assert.Equal(t, "openai", conf.DefaultModel)
	provider, err := conf.Provider()
	require.NoError(t, err)
	assert.Equal(t, "openai", provider)
	model, err := conf.Model()
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o", model, "should fall back to the default model id")
	token, err := conf.Token()
	require.NoError(t, err)
	assert.Equal(t, "sk-test-openai-key", token)
}

func TestRunInit_CreatesConfigFileByProviderName(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy.conf.yml")
	in := strings.NewReader("anthropic\nclaude-sonnet-4-6\nsk-test-anthropic-key\n")
	var out bytes.Buffer

	err := runInit(in, &out, path)

	require.NoError(t, err, "no error expected")
	conf, err := config.YamlConf(path)
	require.NoError(t, err, "config file should be readable")
	provider, err := conf.Provider()
	require.NoError(t, err)
	assert.Equal(t, "anthropic", provider)
	token, err := conf.Token()
	require.NoError(t, err)
	assert.Equal(t, "sk-test-anthropic-key", token)
}

func TestRunInit_GithubKeyIsOptional(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy.conf.yml")
	in := strings.NewReader("2\n\nsk-test-openai-key\n")
	var out bytes.Buffer

	err := runInit(in, &out, path)

	require.NoError(t, err, "no error expected")
	conf, err := config.YamlConf(path)
	require.NoError(t, err, "config file should be readable")
	gh, err := conf.GithubKey()
	require.NoError(t, err)
	assert.Empty(t, gh, "github key should be left empty when skipped")
}

func TestRunInit_GithubKeyIsStoredWhenProvided(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy.conf.yml")
	in := strings.NewReader("2\n\nsk-test-openai-key\ngh-test-key\n")
	var out bytes.Buffer

	err := runInit(in, &out, path)

	require.NoError(t, err, "no error expected")
	conf, err := config.YamlConf(path)
	require.NoError(t, err, "config file should be readable")
	gh, err := conf.GithubKey()
	require.NoError(t, err)
	assert.Equal(t, "gh-test-key", gh)
}

func TestRunInit_ReprompsUntilAPIKeyProvided(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy.conf.yml")
	in := strings.NewReader("1\n\n\nsk-test-deepseek-key\n")
	var out bytes.Buffer

	err := runInit(in, &out, path)

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "this value is required")
	conf, err := config.YamlConf(path)
	require.NoError(t, err, "config file should be readable")
	token, err := conf.Token()
	require.NoError(t, err)
	assert.Equal(t, "sk-test-deepseek-key", token)
}

func TestRunInit_RepromtsOnInvalidProvider(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy.conf.yml")
	in := strings.NewReader("nope\n1\n\nsk-test-deepseek-key\n")
	var out bytes.Buffer

	err := runInit(in, &out, path)

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "invalid choice")
	conf, err := config.YamlConf(path)
	require.NoError(t, err, "config file should be readable")
	provider, err := conf.Provider()
	require.NoError(t, err)
	assert.Equal(t, "deepseek", provider)
}

func TestRunInit_DeclinesOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy.conf.yml")
	require.NoError(t, os.WriteFile(path, []byte("default-model: existing\n"), 0600))
	in := strings.NewReader("n\n")
	var out bytes.Buffer

	err := runInit(in, &out, path)

	require.NoError(t, err, "no error expected")
	assert.Contains(t, out.String(), "aborted")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "default-model: existing\n", string(data), "existing file should be left untouched")
}

func TestRunInit_OverwritesOnConfirm(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".aidy.conf.yml")
	require.NoError(t, os.WriteFile(path, []byte("default-model: existing\n"), 0600))
	in := strings.NewReader("y\n1\n\nsk-test-deepseek-key\n")
	var out bytes.Buffer

	err := runInit(in, &out, path)

	require.NoError(t, err, "no error expected")
	conf, err := config.YamlConf(path)
	require.NoError(t, err, "config file should be readable")
	assert.Equal(t, "deepseek", conf.DefaultModel)
}
