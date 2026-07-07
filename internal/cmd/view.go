package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// viewConfig assembles the config object from the shared flags.
func windowValue(v string) any {
	if v == "" || v == "any" {
		return nil
	}
	return v
}

func viewConfig(tags []string, groupBy string, changedGroup bool, edited string, changedEdited bool) map[string]any {
	config := map[string]any{}
	if tags != nil {
		config["tags"] = tags
	}
	if changedGroup {
		if groupBy == "" || groupBy == "none" {
			config["group_by"] = nil
		} else {
			config["group_by"] = groupBy
		}
	}
	if changedEdited {
		config["edited"] = windowValue(edited)
	}
	return config
}

func createViewCmd() *cobra.Command {
	var (
		name, description, groupBy, edited string
		tags                               []string
	)

	c := &cobra.Command{
		Use:   "create-view",
		Short: "Save a named view of the docs: filter tags + optional group-by.",
		Long: `Create a view — a named, versioned snapshot of the Docs page's
display knobs. group_by a tag scope (like "status") renders a kanban;
no group_by renders the list. Views appear in the Docs page switcher
immediately; pin-view puts one in the sidebar.

The description teaches agents what the view is for — write it like
you write tag descriptions.`,
		Example: `  aveline create-view --name tickets \
    --description "All open work, grouped by status. Check before filing duplicates." \
    --tag ticket --group-by status`,
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
			body := map[string]any{
				"name":        name,
				"description": description,
				"config":      viewConfig(tags, groupBy, cmd.Flags().Changed("group-by"), edited, cmd.Flags().Changed("edited")),
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx, fmt.Sprintf("/api/workspaces/%s/views", ws), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "View name (a slug; shown in the switcher and sidebar).")
	c.Flags().StringVar(&description, "description", "", "What this view is for (agents read this).")
	c.Flags().StringSliceVar(&tags, "tag", nil, "Filter tag (repeatable; empty = all docs).")
	c.Flags().StringVar(&groupBy, "group-by", "", `Tag scope to group by ("status"); omit for a list.`)
	c.Flags().StringVar(&edited, "edited", "", `Only docs last edited within a window ("7d", "24h").`)
	_ = c.MarkFlagRequired("name")
	_ = c.MarkFlagRequired("description")
	return c
}

func editViewCmd() *cobra.Command {
	var (
		newName, description, groupBy, edited string
		tags                                  []string
	)

	c := &cobra.Command{
		Use:   "edit-view <name>",
		Short: "Edit a view (versioned; the screen-level tweaks humans make are never saved — this is).",
		Example: `  aveline edit-view tickets --group-by status
  aveline edit-view tickets --tag ticket --tag urgent
  aveline edit-view tickets --new-name work`,
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
			if description != "" {
				body["description"] = description
			}
			if cmd.Flags().Changed("tag") || cmd.Flags().Changed("group-by") || cmd.Flags().Changed("edited") {
				body["config"] = viewConfig(tags, groupBy, cmd.Flags().Changed("group-by"), edited, cmd.Flags().Changed("edited"))
			}

			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Patch(ctx,
				fmt.Sprintf("/api/workspaces/%s/views/%s", ws, args[0]), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&newName, "new-name", "", "Rename the view.")
	c.Flags().StringVar(&description, "description", "", "New description.")
	c.Flags().StringSliceVar(&tags, "tag", nil, "Replace the filter tag set (repeatable).")
	c.Flags().StringVar(&groupBy, "group-by", "", `New group-by scope; "none" clears it.`)
	c.Flags().StringVar(&edited, "edited", "", `Last-edited window ("7d"); "any" clears it.`)
	return c
}

func listViewsCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-views",
		Short:        "List the workspace's views.",
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
			raw, apiErr := client.Get(ctx, fmt.Sprintf("/api/workspaces/%s/views", ws), nil)
			return handle(raw, apiErr)
		},
	}
}

func deleteViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "delete-view <name>",
		Short:        "Soft-delete a view (restore-view brings it back).",
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
				fmt.Sprintf("/api/workspaces/%s/views/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}

func restoreViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "restore-view <name>",
		Short:        "Restore a soft-deleted view.",
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
				fmt.Sprintf("/api/workspaces/%s/views/%s/restore", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func pinViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "pin-view <name>",
		Short:        "Pin a view to the sidebar.",
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
				fmt.Sprintf("/api/workspaces/%s/views/%s/pin", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func unpinViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "unpin-view <name>",
		Short:        "Remove a view from the sidebar.",
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
				fmt.Sprintf("/api/workspaces/%s/views/%s/pin", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}
