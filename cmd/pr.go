package cmd

import (
	"github.com/spf13/cobra"
)

func newPrCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "pull-request",
		Aliases: []string{"pr"},
		Short:   "Create a PR based on changes in the current branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.PullRequest()
		},
	}
	return command
}
