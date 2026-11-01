package cmd

import (
	"fmt"

	"github.com/aveline-ai/cli/internal/config"
	"github.com/spf13/cobra"
)

func listWorkspacesCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-workspaces",
		Short:        "List workspaces you belong to.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, "/api/workspaces", nil)
			return handle(raw, apiErr)
		},
	}
}

func getWorkspaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get-workspace <slug>",
		Short:        "Show a single workspace by slug.",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, "/api/workspaces/"+args[0], nil)
			return handle(raw, apiErr)
		},
	}
}

func createWorkspaceCmd() *cobra.Command {
	var name, slug, setup string

	c := &cobra.Command{
		Use:   "create-workspace",
		Short: "Create a workspace and add yourself as a member.",
		Long: `Create a workspace and add yourself as a member.

Every new workspace gets an orientation doc. --setup decides what is in
it. "standard" (the default) writes a loop that works on day one and
creates the tags and boards it refers to; "custom" writes the same frame
with nothing filled in, and creates nothing else.

Run ` + "`aveline starter`" + ` first to read exactly what standard would write.
Either way it is the team's doc afterwards, to rewrite as they like.`,
		Example: `  aveline create-workspace --name "Aveline AI"
  aveline create-workspace --name "Aveline AI" --setup custom`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			body := map[string]any{"name": name}
			if slug != "" {
				body["slug"] = slug
			}
			if setup != "" {
				body["setup"] = setup
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx, "/api/workspaces", body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "Workspace display name (required).")
	c.Flags().StringVar(&slug, "slug", "", "Workspace URL slug (optional; derived from name if omitted).")
	c.Flags().StringVar(&setup, "setup", "", `How far to furnish it: "standard" (default) or "custom".`)
	_ = c.MarkFlagRequired("name")
	return c
}

func useWorkspaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use-workspace <slug>",
		Short: "Save a workspace slug as the default for future commands.",
		Long: `Persist <slug> to the config file so future doc / comment / tag
commands target it by default. Equivalent to passing --workspace=<slug>
on every call.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadCfg()
			cfg.Workspace = args[0]
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Printf("{\"ok\":true,\"workspace\":%q}\n", args[0])
			return nil
		},
	}
}
