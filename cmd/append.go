package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newAppendCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "append",
		Aliases: []string{"ap"},
		Short:   "Append all local changes to the last commit",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.AppendCommit(ctx.Git)
		},
	}
	return command
}
