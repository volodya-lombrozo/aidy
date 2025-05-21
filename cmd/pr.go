package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newPrCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "pull-request",
		Aliases: []string{"pr"},
		Short:   "Create a PR based on changes in the current branch",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.PullRequest(ctx.Git, ctx.AI, ctx.GitHub, ctx.Cache, ctx.Output)

		},
	}
	return command
}
