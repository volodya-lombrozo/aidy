package cmd

import (
	"github.com/spf13/cobra"
)

func newRepeatCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "repeat",
		Aliases: []string{"rt"},
		Short:   "Repeats previous command",
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Assistant.Repeat(".aidy_repeat")
		},
	}
	return command
}
