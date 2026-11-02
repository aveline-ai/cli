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
	var name, slug, template string

	c := &cobra.Command{
		Use:   "create-workspace",
		Short: "Create a workspace from a template and add yourself as a member.",
		Long: `Create a workspace and add yourself as a member.

Every workspace starts from a template, so --template is required. The
templates are "autonomous-dev" (recommended: a loop that works on day
one, with the tags and boards it refers to, and how a person hands work
to an agent), "knowledge-base" (durable team knowledge: product,
architecture, briefs, tips, runbooks and oncall, with a tickets board)
and "blank" (the orientation frame with nothing filled in).

` + "`aveline templates`" + ` lists them; ` + "`aveline templates <name>`" + ` shows exactly
what one would create. After creation it is the team's own orientation
doc and tags, to reshape as they like.`,
		Example: `  aveline create-workspace --name "Aveline AI" --template autonomous-dev
  aveline create-workspace --name "Aveline AI" --template blank`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			body := map[string]any{"name": name, "template": template}
			if slug != "" {
				body["slug"] = slug
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx, "/api/workspaces", body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "Workspace display name (required).")
	c.Flags().StringVar(&slug, "slug", "", "Workspace URL slug (optional; derived from name if omitted).")
	c.Flags().StringVar(&template, "template", "", `Template to start from (required): "autonomous-dev", "knowledge-base" or "blank". See `+"`aveline templates`"+`.`)
	_ = c.MarkFlagRequired("name")
	_ = c.MarkFlagRequired("template")
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
