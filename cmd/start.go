package cmd

import (
	"github.com/spf13/cobra"
)

func newStartCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "start [issue]",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(1),
		Short:   "Start a new issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.StartIssue(args[0])
		},
	}
	return command
}
