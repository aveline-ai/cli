package commands

import (
	"fmt"

	"github.com/aveline-ai/cli/internal/config"
	"github.com/aveline-ai/cli/internal/output"
	"github.com/spf13/cobra"
)

func newWorkspaceCmd(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage workspaces",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List workspaces you're a member of",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := g.Client()
			if err != nil {
				return err
			}
			ws, err := c.ListWorkspaces(cmd.Context())
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), map[string]any{"workspaces": ws})
			}
			output.Workspaces(cmd.OutOrStdout(), ws)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "use <slug>",
		Short: "Set the active workspace in the config file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := g.Config()
			if err != nil {
				return err
			}
			cfg.Workspace = args[0]
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "active workspace: %s\n", args[0])
			return nil
		},
	})

	return cmd
}
