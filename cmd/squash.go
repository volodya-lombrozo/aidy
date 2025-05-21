package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newSquashCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "squash",
		Aliases: []string{"sq"},
		Short:   "Squash all commits in the current branch into a single commit",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.Squash(ctx.Git, ctx.AI)
		},
	}
	return command
}
