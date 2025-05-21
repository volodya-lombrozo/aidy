package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newReleaseCmd(ctx *Context) *cobra.Command {
	var repo string
	command := &cobra.Command{
		Use:     "release [increment]",
		Aliases: []string{"r"},
		Args:    cobra.ExactArgs(1),
		Short:   "Create a release based on a semver increment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return aidy.Release(args[0], repo, ctx.Git, ctx.AI, ctx.Output)
		},
	}
	command.Flags().StringVarP(&repo, "repo", "r", "", "repository where to look for tags")
	return command
}
