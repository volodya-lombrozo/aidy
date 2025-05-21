package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newStartCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "start [issue]",
		Aliases: []string{"st"},
		Args:    cobra.ExactArgs(1),
		Short:   "Start a new issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return aidy.StartIssue(args[0], ctx.AI, ctx.Git, ctx.GitHub)
		},
	}
	return command
}
