package integration

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/cmd"
	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/git"
)

const conf = `default-model: 4o-mini

api-keys:
  openai: xxx
  github: yyy
  deepseek: zzz
`

func TestAidyCommit_WithoutRemote(t *testing.T) {
	wd := t.TempDir()
	old, _ := os.Getwd()
	defer func() { _ = os.Chdir(old) }()
	err := os.Chdir(wd)
	require.NoError(t, err, "chdir should work")
	err = os.WriteFile(".aidy.conf.yml", []byte(conf), 0644)
	require.NoError(t, err, "writing file should work")
	g, err := git.NewGit(executor.NewReal())
	require.NoError(t, err, "git should be created without error")
	_, err = g.Run("init")
	require.NoError(t, err, "git init should work")
	_, err = g.Run("config", "user.name", "Your Name")
	require.NoError(t, err, "git config user.name should work")
	_, err = g.Run("config", "user.email", "you@example.com")
	require.NoError(t, err, "git config user.email should work")
	_, err = g.Run("checkout", "-b", "feature")
	require.NoError(t, err, "git checkout should work")
	err = os.WriteFile("file.txt", []byte("content"), 0644)
	require.NoError(t, err, "writing file should work")
	_, err = g.Run("add", "file.txt")
	require.NoError(t, err, "git add should work")
	_, err = g.Run("commit", "-m", "Initial commit")
	require.NoError(t, err, "git commit should work")
	err = os.WriteFile("second-file.txt", []byte("content"), 0644)
	require.NoError(t, err, "writing file should work")
	command := cmd.NewRootCmd(cmd.Real)
	command.SetArgs([]string{"ci", "-n"})
	var out, errb bytes.Buffer
	command.SetOut(&out)
	command.SetErr(&errb)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command.SetContext(ctx)

	err = command.Execute()

	require.NoError(t, err, "command should execute without error")
}
