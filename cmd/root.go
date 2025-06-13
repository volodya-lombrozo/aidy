package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

type Context struct {
	Assistant aidy.Aidy
}

func Execute() error {
	return newRootCmd(real).Execute()
}

func real(summary, aider, ailess, silent, debug bool) aidy.Aidy {
	return aidy.NewAidy(summary, aider, ailess, silent, debug)
}

func newRootCmd(create func(bool, bool, bool, bool, bool) aidy.Aidy) *cobra.Command {
	var ctx Context
	var ailess bool
	var aider bool
	var summary bool
	var silent bool
	var debug bool
	root := &cobra.Command{
		Use:   "aidy",
		Short: "aidy - ai-powered github cli helper",
		Long:  "Aidy assists you with generating commit messages, pull requests, issues, and releases",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx.Assistant = create(summary, aider, ailess, silent, debug)
		},
	}
	root.PersistentFlags().BoolVarP(&ailess, "no-ai", "n", false, "don't use AI")
	root.PersistentFlags().BoolVarP(&summary, "summary", "s", false, "use a project summary in AI requests")
	root.PersistentFlags().BoolVar(&aider, "aider", false, "use aider configuration")
	root.PersistentFlags().BoolVarP(&silent, "quiet", "q", false, "be silent, don't print logs")
	root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "print debug logs")
	root.AddCommand(
		newCommitCmd(&ctx),
		newIssueCmd(&ctx),
		newReleaseCmd(&ctx),
		newPrCmd(&ctx),
		newHealCmd(&ctx),
		newSquashCmd(&ctx),
		newAppendCmd(&ctx),
		newConfigCmd(&ctx),
		newCleanCmd(&ctx),
		newStartCmd(&ctx),
		newDiffCmd(&ctx),
	)
	return root
}
