package cmd

import (
	"github.com/spf13/cobra"
)

func newCommitCmd(ctx *Context) *cobra.Command {
	return &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "Make a commit with AI-generated message",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctx.Assistant.Commit()
		},
	}
}
