package commands

import (
	"github.com/aveline-ai/cli/internal/output"
	"github.com/spf13/cobra"
)

func newWhoamiCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the authenticated user and their workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := g.Client()
			if err != nil {
				return err
			}
			me, err := c.Me(cmd.Context())
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), me)
			}
			output.Me(cmd.OutOrStdout(), me)
			return nil
		},
	}
}
