package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func createDataSourceCmd() *cobra.Command {
	var name, url, password string

	c := &cobra.Command{
		Use:   "create-data-source",
		Short: "Connect an external database for chart blocks.",
		Long: `Connect an external database this workspace can chart from.

--url is a connection-string TEMPLATE: your normal postgres:// or
mysql:// URL with the literal placeholder <password> where the
password goes. The template contains no secret, so every read surface
shows it verbatim — you can always see exactly where a source points.

--password is the secret. It is encrypted at rest, has no read path
(not even for you — lose it and you rotate), and is substituted into
the template only at query time. Use a READ-ONLY database user:
queries are forced read-only server-side, but least privilege is
yours to grant.

Chart blocks reference the source by name:
    {"type": "chart", "source": "<name>", "query": "select ...",
     "viz": {"type": "line", "x": "day", "y": "signups"}}`,
		Example: `  aveline create-data-source --name prod \
    --url "postgres://metrics_ro:<password>@db.internal:5432/app" \
    --password "s3cret"`,
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
				map[string]any{"name": name, "url": url, "password": password})
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "Source name (a slug; chart blocks reference it).")
	c.Flags().StringVar(&url, "url", "", "Connection template with a literal <password> placeholder.")
	c.Flags().StringVar(&password, "password", "", "The secret. Stored encrypted, never readable again.")
	_ = c.MarkFlagRequired("name")
	_ = c.MarkFlagRequired("url")
	_ = c.MarkFlagRequired("password")
	return c
}

func editDataSourceCmd() *cobra.Command {
	var newName, url, password string

	c := &cobra.Command{
		Use:   "edit-data-source <name>",
		Short: "Rename, rotate the password, or repoint a data source (versioned).",
		Long: `Edit a data source. Mints a new version on the same identity, so
chart blocks referencing it never notice. The superseded version's
password is destroyed in the same transaction.

Rules:
  --password alone         rotate the secret (most common; always safe)
  --new-name alone         rename (doesn't touch the connection)
  --url ...                requires --password with it: a stored secret
                           is never combined with connection settings
                           it wasn't written with.`,
		Example: `  aveline edit-data-source prod --password "new-s3cret"
  aveline edit-data-source prod --new-name analytics
  aveline edit-data-source prod \
    --url "postgres://metrics_ro:<password>@new-host:5432/app" --password "s3cret"`,
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

			body := map[string]any{}
			if newName != "" {
				body["new_name"] = newName
			}
			if url != "" {
				body["url"] = url
			}
			if cmd.Flags().Changed("password") {
				body["password"] = password
			}

			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Patch(ctx,
				fmt.Sprintf("/api/workspaces/%s/data-sources/%s", ws, args[0]), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&newName, "new-name", "", "New source name.")
	c.Flags().StringVar(&url, "url", "", "New connection template (requires --password).")
	c.Flags().StringVar(&password, "password", "", "New secret (alone = rotation).")
	return c
}

func listDataSourcesCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-data-sources",
		Short:        "List the workspace's data sources (templates only, never secrets).",
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
		Short: "Delete a data source; its password is destroyed immediately.",
		Long: `Delete a data source. The row survives for audit (template, adapter,
who connected it, when) but the encrypted password is hard-deleted in
the same operation and cannot be recovered. Charts referencing it show
an error state. There is no restore: connect a new source.`,
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
