package cmd

import (
	"github.com/spf13/cobra"
)

func newCleanCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cl"},
		Short:   "Clean the aidy cache",
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Assistant.Clean()
		},
	}
	return command
}
