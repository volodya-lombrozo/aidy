package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newCommitCmd(ctx *Context) *cobra.Command {
	return &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "Make a commit with AI-generated message",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.Commit(ctx.Git, ctx.AI)
		},
	}
}
