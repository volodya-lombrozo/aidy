package cmd

import (
	"github.com/spf13/cobra"
)

func newHealCmd(ctx *Context) *cobra.Command {
	command := &cobra.Command{
		Use:     "heal",
		Aliases: []string{"hl"},
		Short:   "Fix the current commit message if the AI made mistakes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.Heal()
		},
	}
	return command
}
