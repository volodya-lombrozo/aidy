package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newConfigCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "config",
		Aliases: []string{"conf"},
		Short:   "Print the current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.Pconfig(ctx.Config)
		},
	}
	return command
}
