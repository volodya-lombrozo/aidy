package aidy

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volodya-lombrozo/aidy/internal/ai"
	"github.com/volodya-lombrozo/aidy/internal/cache"
	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/git"
	"github.com/volodya-lombrozo/aidy/internal/github"
	"github.com/volodya-lombrozo/aidy/internal/gitlab"
	"github.com/volodya-lombrozo/aidy/internal/log"
	"github.com/volodya-lombrozo/aidy/internal/output"
)

func TestParseRepository(t *testing.T) {
	for _, test := range []struct {
		remote string
		want   string
	}{
		{remote: "git@code.company.com:group/subgroup/project.git", want: "https://code.company.com/group/subgroup/project"},
		{remote: "ssh://git@code.company.com:2222/group/project.git", want: "https://code.company.com/group/project"},
		{remote: "https://user:secret@code.company.com:8443/group/project.git", want: "https://code.company.com:8443/group/project"},
		{remote: "origin https://CODE.company.com/group/project.git (fetch)", want: "https://code.company.com/group/project"},
		{remote: "/local/group/project"},
		{remote: "file:///local/group/project"},
		{remote: "https://code.company.com/project"},
		{remote: "https://code.company.com/group/project?token=secret"},
		{remote: "https://code.company.com/group/project#fragment"},
	} {
		t.Run(test.remote, func(t *testing.T) {
			repo, ok := parseRepository(test.remote)

			assert.Equal(t, test.want != "", ok)
			assert.Equal(t, test.want, repo.url)
		})
	}
}

func TestStartIssueSamePathAcrossHosts(t *testing.T) {
	for _, test := range []struct {
		choice string
		target string
		github int
	}{
		{choice: "1", target: "https://code.company.com/group/project", github: 0},
		{choice: "2", target: "https://github.com/group/project", github: 1},
	} {
		t.Run(test.target, func(t *testing.T) {
			shell := executor.NewMock()
			shell.Output = "origin git@github.com:group/project.git (fetch)\nmirror git@code.company.com:group/project.git (fetch)\nmirror git@code.company.com:group/project.git (push)"
			stored := cache.NewMockAidyCache()
			stored.WithRemote("")
			reader, writer, err := os.Pipe()
			require.NoError(t, err)
			t.Cleanup(func() { _ = reader.Close() })
			_, err = writer.WriteString(test.choice + "\n")
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			printer := output.NewMock()
			gh := github.NewMock()
			brain := &issueBranchAI{AI: ai.NewMockAI()}
			assistant := &real{git: git.NewMockWithShell(shell), github: gh, gitlab: gitlab.NewMock(), ai: brain, cache: stored, logger: log.Default(), printer: printer, in: reader}

			require.NoError(t, assistant.StartIssue("11"))

			assert.Equal(t, test.target, stored.Remote())
			assert.Equal(t, test.github, gh.Descriptions)
			assert.Contains(t, printer.Captured(), "https://github.com/group/project")
			assert.Contains(t, printer.Captured(), "https://code.company.com/group/project")
			if test.github == 0 {
				assert.Contains(t, brain.description, test.target)
			}
			require.NoError(t, assistant.StartIssue("12"))
			assert.Equal(t, test.github*2, gh.Descriptions)
			assert.Equal(t, test.target, stored.Remote())
		})
	}
}

func TestStartIssueAmbiguousLegacyTarget(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "origin https://github.com/group/project.git (fetch)\nmirror https://code.company.com/group/project.git (fetch)"
	stored := cache.NewMockAidyCache()
	stored.WithRemote("group/project")
	gh := github.NewMock()
	assistant := &real{git: git.NewMockWithShell(shell), github: gh, gitlab: gitlab.NewMock(), cache: stored, logger: log.Default()}

	err := assistant.StartIssue("11")

	require.ErrorContains(t, err, "ambiguous target repository")
	assert.Contains(t, err.Error(), "aidy clean")
	assert.Zero(t, gh.Descriptions)
	assert.NotContains(t, strings.Join(shell.Commands, "\n"), "checkout")
}

func TestStartIssueMigratesLegacyGitlabTarget(t *testing.T) {
	shell := executor.NewMock()
	shell.Output = "origin git@code.company.com:group/project.git (fetch)"
	stored := cache.NewMockAidyCache()
	stored.WithRemote("group/project")
	gh := github.NewMock()
	assistant := &real{git: git.NewMockWithShell(shell), github: gh, gitlab: gitlab.NewMock(), ai: ai.NewMockAI(), cache: stored, logger: log.Default()}

	require.NoError(t, assistant.StartIssue("11"))

	assert.Equal(t, "https://code.company.com/group/project", stored.Remote())
	assert.Zero(t, gh.Descriptions)
}
