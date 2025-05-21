package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newDiffCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "diff",
		Aliases: []string{"df"},
		Short:   "Print the current diff that will be used to generate the commit message",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.Diff(ctx.Git)
		},
	}
	return command
}
