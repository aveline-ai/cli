package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func createQueryCmd() *cobra.Command {
	var name, source, sql, description string

	c := &cobra.Command{
		Use:   "create-query",
		Short: "Create a named catalog query (charts reference these).",
		Long: `Create a named, versioned query — the reusable unit charts are built
on. Two kinds:

  RAW      pass --source <data-source-name>. The SQL runs against that
           external database in its own dialect.
  DERIVED  omit --source. The SQL is the analytics dialect (DuckDB) over
           other catalog queries by name — regressions, window
           functions, and joins across sources the source dialects
           can't do. Derived queries can build on other derived queries.

--name is a table-safe identifier (lowercase letter first, then
letters/digits/underscores): it becomes a table name inside other
queries' SQL.

--sql accepts text directly, '-' for stdin, or @PATH for a file.

Charts reference a query by name:
    {"type": "chart", "query_ref": "<name>",
     "viz": {"type": "line", "x": "day", "y": "signups"}}`,
		Example: `  aveline create-query --name signups --source prod \
    --sql "select created_at::date as day, count(*) as n from users group by 1"

  # derived: join two catalog queries in the engine
  aveline create-query --name cac \
    --sql "select s.day, spend.amt / s.n as cac from signups s join spend using (day)"`,
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
			text, err := readTextInput(sql)
			if err != nil {
				return err
			}
			body := map[string]any{"name": name, "sql": text}
			if source != "" {
				body["source"] = source
			}
			if description != "" {
				body["description"] = description
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/queries", ws), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "Query name (a table-safe identifier).")
	c.Flags().StringVar(&source, "source", "", "Data source name for a RAW query; omit for DERIVED.")
	c.Flags().StringVar(&sql, "sql", "", "The SQL. Text, '-' for stdin, or @PATH.")
	c.Flags().StringVar(&description, "description", "", "One line on what this query answers. Shown wherever the query is listed; give every query one.")
	_ = c.MarkFlagRequired("name")
	_ = c.MarkFlagRequired("sql")
	return c
}

func editQueryCmd() *cobra.Command {
	var newName, sql, description string

	c := &cobra.Command{
		Use:   "edit-query <name>",
		Short: "Rename or rewrite a catalog query (versioned).",
		Long: `Edit a catalog query. Mints a new version on the same identity, so
charts referencing it by name never notice.

  --sql ...       new SQL (text, '-' for stdin, or @PATH)
  --new-name ...  rename. Rejected while other DERIVED queries reference
                  the old name — update them first. (Charts reference by
                  name too and will show an error card until updated.)`,
		Example: `  aveline edit-query signups --sql @signups.sql
  aveline edit-query signups --new-name daily_signups`,
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
			if cmd.Flags().Changed("sql") {
				text, err := readTextInput(sql)
				if err != nil {
					return err
				}
				body["sql"] = text
			}
			if cmd.Flags().Changed("description") {
				body["description"] = description
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Patch(ctx,
				fmt.Sprintf("/api/workspaces/%s/queries/%s", ws, args[0]), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&newName, "new-name", "", "New query name.")
	c.Flags().StringVar(&sql, "sql", "", "New SQL. Text, '-' for stdin, or @PATH.")
	c.Flags().StringVar(&description, "description", "", "New description; pass \"\" to clear.")
	return c
}

func listQueriesCmd() *cobra.Command {
	var source string

	c := &cobra.Command{
		Use:   "list-queries",
		Short: "List catalog queries (optionally those built on one source).",
		Long: `List the workspace's catalog queries. Pass --source <name> for the
lineage of one data source: the queries built on it (raw queries for an
external source; the derived catalog for the built-in "derived" source).`,
		Example: `  aveline list-queries
  aveline list-queries --source prod`,
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
			var q url.Values
			if source != "" {
				q = url.Values{"source": {source}}
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx,
				fmt.Sprintf("/api/workspaces/%s/queries", ws), q)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&source, "source", "", "Only queries built on this data source.")
	return c
}

func getQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get-query <name>",
		Short:        "Show a catalog query (its SQL, kind, source, version).",
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
			raw, apiErr := client.Get(ctx,
				fmt.Sprintf("/api/workspaces/%s/queries/%s", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func deleteQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-query <name>",
		Short: "Soft-delete a catalog query (restore-query brings it back).",
		Long: `Soft-delete a catalog query. Rejected while other DERIVED queries
reference it — update them first. Charts referencing it show an error
card. Unlike data sources, queries hold no secret, so this is
reversible: restore-query.`,
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
				fmt.Sprintf("/api/workspaces/%s/queries/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}

func restoreQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "restore-query <name>",
		Short:        "Restore a soft-deleted catalog query.",
		Long:         `Restore a soft-deleted query. Fails if the name was taken since.`,
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
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/queries/%s/restore", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}
