package cmd

import (
	"github.com/spf13/cobra"
)

func newSquashCmd(ctx *Context) *cobra.Command {
	var issue bool
	command := &cobra.Command{
		Use:     "squash",
		Aliases: []string{"sq"},
		Short:   "Squash all commits in the current branch into a single commit",
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Assistant.Squash(issue)
		},
	}
	command.Flags().BoolVarP(&issue, "issue", "i", false, "use issue description to genearate commit message")
	return command
}
