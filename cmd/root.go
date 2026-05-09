package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

type Context struct {
	Assistant aidy.Aidy
}

func Execute() error {
	return NewRootCmd(Real).Execute()
}

func Real(summary, aider, ailess, silent, debug bool, language string) aidy.Aidy {
	return aidy.NewAidy(summary, aider, ailess, silent, debug, language)
}

func NewRootCmd(create func(bool, bool, bool, bool, bool, string) aidy.Aidy) *cobra.Command {
	var ctx Context
	var ailess bool
	var aider bool
	var summary bool
	var silent bool
	var debug bool
	var language string
	root := &cobra.Command{
		Use:     "aidy",
		Short:   "aidy - ai-powered github cli helper",
		Long:    "Aidy assists you with generating commit messages, pull requests, issues, and releases",
		Version: Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx.Assistant = create(summary, aider, ailess, silent, debug, language)
		},
	}
	root.PersistentFlags().BoolVarP(&ailess, "no-ai", "n", false, "don't use AI")
	root.PersistentFlags().BoolVarP(&summary, "summary", "s", false, "use a project summary in AI requests")
	root.PersistentFlags().BoolVar(&aider, "aider", false, "use aider configuration")
	root.PersistentFlags().BoolVarP(&silent, "quiet", "q", false, "be silent, don't print logs")
	root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "print debug logs")
	root.PersistentFlags().StringVarP(&language, "language", "l", "en", "language for AI-generated text (e.g. fr, de, ja)")
	root.AddCommand(
		newCommitCmd(&ctx),
		newIssueCmd(&ctx),
		newReleaseCmd(&ctx),
		newPrCmd(&ctx),
		newMrCmd(&ctx),
		newHealCmd(&ctx),
		newSquashCmd(&ctx),
		newAppendCmd(&ctx),
		newConfigCmd(&ctx),
		newCleanCmd(&ctx),
		newStartCmd(&ctx),
		newDiffCmd(&ctx),
		newVersionCmd(),
	)
	return root
}
