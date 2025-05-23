package cmd

import (
	"github.com/spf13/cobra"
)

func newConfigCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "config",
		Aliases: []string{"conf"},
		Short:   "Print the current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Assistant.PrintConfig()
		},
	}
	return command
}
