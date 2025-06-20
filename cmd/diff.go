package cmd

import (
	"github.com/spf13/cobra"
)

func newDiffCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "diff",
		Aliases: []string{"df"},
		Short:   "Print the current diff that will be used to generate the commit message",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.Diff()
		},
	}
	return command
}
