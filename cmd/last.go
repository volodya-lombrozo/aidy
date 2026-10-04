package cmd

import (
	"github.com/spf13/cobra"
)

func newLastCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:   "last",
		Short: "Repeat the last generated command, including your edits",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.Last()
		},
	}
	return command
}
