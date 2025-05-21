package cmd

import (
	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
)

func newHealCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "heal",
		Aliases: []string{"hl"},
		Short:   "Fix the current commit message if the AI made mistakes",
		Run: func(cmd *cobra.Command, args []string) {
			aidy.Heal(ctx.Git)
		},
	}
	return command
}
