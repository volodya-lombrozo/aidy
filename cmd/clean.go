package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newCleanCmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cl"},
		Short:   "Clean the aidy cache",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.Clean()
		},
	}
	return command
}
