package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func windowValue(v string) any {
	if v == "" || v == "any" {
		return nil
	}
	return v
}

func scopeValue(v string) any {
	if v == "" || v == "none" {
		return nil
	}
	return v
}

// viewKnobs holds the config flags create-view and edit-view share.
type viewKnobs struct {
	tags                                []string
	groupBy, subGroupBy, edited, layout string
}

var viewKnobFlags = []string{"tag", "group-by", "sub-group-by", "edited", "layout"}

func (k *viewKnobs) bind(f *pflag.FlagSet, editing bool) {
	if editing {
		f.StringSliceVar(&k.tags, "tag", nil, "Replace the filter tag set (repeatable).")
		f.StringVar(&k.groupBy, "group-by", "", `New group-by scope; "none" clears it.`)
		f.StringVar(&k.subGroupBy, "sub-group-by", "", `Sub-group scope; "none" clears it.`)
		f.StringVar(&k.edited, "edited", "", `Last-edited window ("7d"); "any" clears it.`)
		f.StringVar(&k.layout, "layout", "", `How groups are laid out: "list" (stacked sections) or "board" (columns side by side).`)
		return
	}
	f.StringSliceVar(&k.tags, "tag", nil, "Filter tag (repeatable; empty = all docs).")
	f.StringVar(&k.groupBy, "group-by", "", `Tag scope to group by ("status"); omit for a list.`)
	f.StringVar(&k.subGroupBy, "sub-group-by", "", `Second scope for subsections (needs --group-by).`)
	f.StringVar(&k.edited, "edited", "", `Only docs last edited within a window ("7d", "24h").`)
	f.StringVar(&k.layout, "layout", "", `How groups are laid out: "list" (stacked sections, the default) or "board" (columns side by side, a kanban). Needs --group-by.`)
}

// changed reports whether any config flag was given.
func (k *viewKnobs) changed(f *pflag.FlagSet) bool {
	for _, name := range viewKnobFlags {
		if f.Changed(name) {
			return true
		}
	}
	return false
}

// config assembles the config object from the flags that were given, so
// an edit only touches the keys named on the command line.
func (k *viewKnobs) config(f *pflag.FlagSet) map[string]any {
	config := map[string]any{}
	if k.tags != nil {
		config["tags"] = k.tags
	}
	if f.Changed("group-by") {
		config["group_by"] = scopeValue(k.groupBy)
	}
	if f.Changed("sub-group-by") {
		config["sub_group_by"] = scopeValue(k.subGroupBy)
	}
	if f.Changed("edited") {
		config["edited"] = windowValue(k.edited)
	}
	if f.Changed("layout") {
		config["layout"] = k.layout
	}
	return config
}

func createViewCmd() *cobra.Command {
	var (
		name, description, bucket string
		knobs                     viewKnobs
	)

	c := &cobra.Command{
		Use:   "create-view",
		Short: "Save a named view of the docs: filter tags + optional group-by.",
		Long: `Create a view — a named, versioned snapshot of the Docs page's
display knobs. group_by a tag scope (like "status") splits the docs
into one section per scope value; --layout board lays those sections
out side by side as a kanban. No group_by renders the list. Views
appear in the Docs page switcher immediately; pin-view puts one in
the sidebar.

The description teaches agents what the view is for — write it like
you write tag descriptions.`,
		Example: `  aveline create-view --name tickets \
    --description "All open work, grouped by status. Check before filing duplicates." \
    --tag ticket --group-by status
  aveline create-view --name board \
    --description "Tickets as a kanban: one column per status." \
    --tag ticket --group-by status --layout board`,
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
				"config":      knobs.config(cmd.Flags()),
			}
			if bucket != "" {
				body["bucket"] = bucket
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx, fmt.Sprintf("/api/workspaces/%s/views", ws), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "View name (a slug; shown in the switcher and sidebar).")
	c.Flags().StringVar(&description, "description", "", "What this view is for (agents read this).")
	knobs.bind(c.Flags(), false)
	c.Flags().StringVar(&bucket, "bucket", "", `Bucket to create the view in: "team", a project bucket name, or omit for yours (private by default).`)
	_ = c.MarkFlagRequired("name")
	_ = c.MarkFlagRequired("description")
	return c
}

func editViewCmd() *cobra.Command {
	var (
		newName, description string
		knobs                viewKnobs
	)

	c := &cobra.Command{
		Use:   "edit-view <name>",
		Short: "Edit a view (versioned; the screen-level tweaks humans make are never saved — this is).",
		Example: `  aveline edit-view tickets --group-by status
  aveline edit-view tickets --layout board
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
			if knobs.changed(cmd.Flags()) {
				body["config"] = knobs.config(cmd.Flags())
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
	knobs.bind(c.Flags(), true)
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
