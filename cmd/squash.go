package cmd

import (
	"github.com/spf13/cobra"
)

func newSquashCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "squash",
		Aliases: []string{"sq"},
		Short:   "Squash all commits in the current branch into a single commit",
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Assistant.Squash()
		},
	}
	return command
}
