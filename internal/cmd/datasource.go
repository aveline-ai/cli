package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func createDataSourceCmd() *cobra.Command {
	var name, url string

	c := &cobra.Command{
		Use:   "create-data-source",
		Short: "Connect an external database for chart blocks.",
		Long: `Connect an external database this workspace can chart from.

The adapter is derived from the URL scheme: postgres:// (or
postgresql://) and mysql:// are supported. Use a READ-ONLY database
user: chart queries are forced read-only server-side, but least
privilege is yours to grant.

The URL is encrypted at rest and write-only: no API response ever
echoes it back. Reads show name, adapter, host, and database only.

Chart blocks reference the source by name:
    {"type": "chart", "source": "<name>", "query": "select ...",
     "viz": {"type": "line", "x": "day", "y": "signups"}}`,
		Example: `  aveline create-data-source --name prod \
    --url "postgres://metrics_ro:PASS@db.internal:5432/app"`,
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
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/data-sources", ws),
				map[string]any{"name": name, "url": url})
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "Source name (a slug; chart blocks reference it).")
	c.Flags().StringVar(&url, "url", "", "Connection URL (postgres:// or mysql://). Stored encrypted, never echoed.")
	_ = c.MarkFlagRequired("name")
	_ = c.MarkFlagRequired("url")
	return c
}

func listDataSourcesCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-data-sources",
		Short:        "List the workspace's data sources (never their credentials).",
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
				fmt.Sprintf("/api/workspaces/%s/data-sources", ws), nil)
			return handle(raw, apiErr)
		},
	}
}

func deleteDataSourceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-data-source <name>",
		Short: "Delete a data source; its credential is destroyed immediately.",
		Long: `Delete a data source. The row survives for audit (name, adapter,
who connected it, when) but the encrypted credential is hard-deleted
in the same operation and cannot be recovered. Charts referencing it
show an error state. There is no restore: connect a new source.`,
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
				fmt.Sprintf("/api/workspaces/%s/data-sources/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}
