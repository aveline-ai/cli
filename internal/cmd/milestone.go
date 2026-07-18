package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func createMilestoneCmd() *cobra.Command {
	var name, date, description string

	c := &cobra.Command{
		Use:   "create-milestone",
		Short: "Record a dated milestone; time-series charts in range auto-annotate.",
		Long: `Record a timeline milestone — a dated workspace fact (a release, a
pricing change, a migration). Every time-series chart whose x-range
spans the date draws it as a labeled vertical marker, so "did the spike
follow the release?" answers itself.

Milestones are data, not chart config: recorded once, drawn on every
chart at once. Call this from a deploy pipeline to annotate releases
automatically.`,
		Example: `  aveline create-milestone --name "v1.4 shipped" --date 2026-07-06
  aveline create-milestone --name "pricing change" --date 2026-07-10 \
    --description "Pro tier moved to $12/seat"`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ws, err := resolveWorkspaceSlug()
			if err != nil {
				return err
			}
			body := map[string]any{"name": name, "date": date}
			if description != "" {
				body["description"] = description
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/milestones", ws), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "What happened, e.g. \"v1.4 shipped\".")
	c.Flags().StringVar(&date, "date", "", "When (YYYY-MM-DD).")
	c.Flags().StringVar(&description, "description", "", "Optional detail shown on hover and in listings.")
	_ = c.MarkFlagRequired("name")
	_ = c.MarkFlagRequired("date")
	return c
}

func listMilestonesCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-milestones",
		Short:        "List the workspace's timeline milestones, oldest first.",
		Example:      `  aveline list-milestones`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ws, err := resolveWorkspaceSlug()
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx,
				fmt.Sprintf("/api/workspaces/%s/milestones", ws), nil)
			return handle(raw, apiErr)
		},
	}
}

func deleteMilestoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "delete-milestone <id>",
		Short:        "Delete a milestone by id (see list-milestones); charts stop drawing it.",
		Example:      `  aveline delete-milestone 6e5f6f0e-...`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ws, err := resolveWorkspaceSlug()
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Delete(ctx,
				fmt.Sprintf("/api/workspaces/%s/milestones/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}
