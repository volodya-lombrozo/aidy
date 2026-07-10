package cmd

import (
	"github.com/spf13/cobra"
)

func newMrCmd(ctx *Context) *cobra.Command {
	fixes := false
	target := ""
	duplicate := false
	source := ""
	command := &cobra.Command{
		Use:     "merge-request",
		Aliases: []string{"mr"},
		Short:   "Create a MR based on changes in the current branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.MergeRequest(fixes, target, duplicate, source)
		},
	}
	command.Flags().BoolVarP(&fixes, "fixes", "f", false, "Create a MR with 'fixes' keyword")
	command.Flags().StringVarP(&target, "target", "t", "", "Target branch for the MR")
	command.Flags().BoolVar(&duplicate, "duplicate", false, "Reuse an existing MR's title and body against --target")
	command.Flags().StringVar(&source, "source", "", "Branch whose existing MR to duplicate (defaults to the current branch)")
	return command
}
