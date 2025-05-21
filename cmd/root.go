package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/ai"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
	"github.com/volodya-lombrozo/aidy/internal/cache"
	"github.com/volodya-lombrozo/aidy/internal/config"
	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/git"
	"github.com/volodya-lombrozo/aidy/internal/github"
	"github.com/volodya-lombrozo/aidy/internal/output"
)

type Context struct {
	Git    git.Git
	GitHub github.Github
	AI     ai.AI
	Output output.Output
	Config config.Config
	Cache  cache.AidyCache
}

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	var ctx Context
	var ailess bool
	var aider bool
	var summary bool
	root := &cobra.Command{
		Use:   "aidy",
		Short: "aidy - ai-powered github cli helper",
		Long:  "Aidy assists you with generating commit messages, pull requests, issues, and releases",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			shell := executor.NewRealExecutor()
			out := output.NewEditor(shell)
			ctx.Output = out
			ctx.Git = git.NewGit(shell)
			aidy.CheckGitInstalled(ctx.Git)
			ctx.Cache = newcache(ctx.Git)
			conf := newfonfig(aider, ctx.Git)
			ctx.AI = brain(ailess, summary, conf)
			ctx.GitHub = newgithub(ctx.Git, conf, ctx.Cache)
			aidy.InitSummary(summary, ctx.AI, ctx.Cache)
		},
	}
	root.PersistentFlags().BoolVarP(&ailess, "no-ai", "n", false, "don't use ai")
	root.PersistentFlags().BoolVarP(&summary, "summary", "s", false, "use a project summary in ai requests")
	root.PersistentFlags().BoolVar(&aider, "aider", false, "use aider configuration")
	root.AddCommand(
		newCommitCmd(&ctx),
		newIssueCmd(&ctx),
		newReleaseCmd(&ctx),
		newPrCmd(&ctx),
		newHealCmd(&ctx),
		newSquashCmd(&ctx),
		newAppendCmd(&ctx),
		newConfigCmd(&ctx),
		newCleanCmd(),
		newStartCmd(&ctx),
		newDiffCmd(&ctx),
	)
	return root
}

func newgithub(git git.Git, conf config.Config, cache cache.AidyCache) github.Github {
	token, err := conf.GetGithubAPIKey()
	if err != nil {
		log.Fatalf("error getting github token: %v", err)
	}
	return github.NewGithub("https://api.github.com", git, token, cache)
}

func newfonfig(aider bool, git git.Git) config.Config {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home directory: %v", err)
	}
	var conf config.Config
	if aider {
		conf = config.NewAiderConf(fmt.Sprintf("%s/.aider.conf.yml", home))
	} else {
		conf = config.NewCascadeConfig(git)
	}
	return conf
}

func brain(ailess bool, sumrequired bool, conf config.Config) ai.AI {
	if ailess {
		return ai.NewMockAI()
	}
	model, err := conf.GetModel()
	if err != nil {
		log.Fatalf("Can't find GitHub token in configuration")
	}
	var brain ai.AI
	if model == "deepseek-chat" {
		apiKey, err := conf.GetDeepseekAPIKey()
		if err != nil {
			log.Fatalf("Error getting Deepseek API key: %v", err)
		}
		if apiKey == "" {
			log.Fatalf("Deepseek API key not found in config file")
		} else {
			log.Println("Deepseek key is found")
		}
		brain = ai.NewDeepSeekAI(apiKey, sumrequired)
	} else {
		apiKey, err := conf.GetOpenAIAPIKey()
		if err != nil {
			log.Fatalf("Error getting OpenAI API key: %v", err)
		}
		if apiKey == "" {
			log.Fatalf("OpenAI API key not found in config file")
		}
		brain = ai.NewOpenAI(apiKey, model, 0.2, sumrequired)
	}
	return brain
}

func newcache(repo git.Git) cache.AidyCache {
	gitcache, err := cache.NewGitCache(".aidy/cache.js", repo)
	if err != nil {
		log.Fatalf("Can't open cache %v", err)
	}
	return cache.NewAidyCache(gitcache)
}
