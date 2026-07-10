package cmd

import (
	"github.com/spf13/cobra"
)

func newPrCmd(ctx *Context) *cobra.Command {
	fixes := false
	target := ""
	duplicate := false
	source := ""
	command := &cobra.Command{
		Use:     "pull-request",
		Aliases: []string{"pr"},
		Short:   "Create a PR based on changes in the current branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.PullRequest(fixes, target, duplicate, source)
		},
	}
	command.Flags().BoolVarP(&fixes, "fixes", "f", false, "Create a PR with 'fixes' keyword")
	command.Flags().StringVarP(&target, "target", "t", "", "Target branch for the PR")
	command.Flags().BoolVar(&duplicate, "duplicate", false, "Reuse an existing PR's title and body against --target")
	command.Flags().StringVar(&source, "source", "", "Branch whose existing PR to duplicate (defaults to the current branch)")
	return command
}
