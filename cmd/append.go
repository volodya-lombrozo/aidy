package cmd

import (
	"github.com/spf13/cobra"
)

func newAppendCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "append",
		Aliases: []string{"ap"},
		Short:   "Append all local changes to the last commit",
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Assistant.Append()
		},
	}
	return command
}
