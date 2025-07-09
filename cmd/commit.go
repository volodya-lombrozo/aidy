package cmd

import (
	"github.com/spf13/cobra"
)

func newCommitCmd(ctx *Context) *cobra.Command {
	var issue bool 
	command := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "Make a commit with AI-generated message",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.Commit(issue)
		},
	}
	command.Flags().BoolVarP(&issue, "issue", "i", false, "use issue description to genearate commit message")
	return command
}
