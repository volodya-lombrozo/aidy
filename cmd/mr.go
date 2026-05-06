package cmd

import (
	"github.com/spf13/cobra"
)

func newMrCmd(ctx *Context) *cobra.Command {
	fixes := false
	command := &cobra.Command{
		Use:     "merge-request",
		Aliases: []string{"mr"},
		Short:   "Create a MR based on changes in the current branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.MergeRequest(fixes)
		},
	}
	command.Flags().BoolVarP(&fixes, "fixes", "f", false, "Create a MR with 'fixes' keyword")
	return command
}
