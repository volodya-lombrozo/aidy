package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newIssueCmd(ctx *Context) *cobra.Command {
	return &cobra.Command{
		Use:     "issue [description]",
		Aliases: []string{"i"},
		Args:    cobra.ExactArgs(1),
		Short:   "Generate a GitHub issue using an AI-generated title, body, and labels",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.Issue(args[0], ctx.AI, ctx.GitHub, ctx.Cache, ctx.Output)
		},
	}
}
