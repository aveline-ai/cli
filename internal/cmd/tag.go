package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func listTagsCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-tags",
		Short:        "List all tags in the current workspace (with usage counts).",
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
				fmt.Sprintf("/api/workspaces/%s/tags", ws), nil)
			return handle(raw, apiErr)
		},
	}
}

func getTagCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get-tag <tag-slug>",
		Short:        "Show a single tag's metadata.",
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
				fmt.Sprintf("/api/workspaces/%s/tags/%s", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func createTagCmd() *cobra.Command {
	var name, description string

	c := &cobra.Command{
		Use:   "create-tag",
		Short: "Create a new tag in the current workspace.",
		Long: `Tags are workspace-scoped — they must exist before being attached
to a doc. Slug rules: lowercase letters, digits, hyphens, must
start with a letter or digit, 1–40 chars.

--description is required; it feeds search and helps humans + LLMs
understand what the tag covers.`,
		Example:      `  aveline create-tag --name runbook --description "Operational runbooks."`,
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
			// Server canonical field is `slug`; `name` is the CLI-facing
			// alias. Map here so the contract stays clean on the wire.
			body := map[string]any{
				"slug":        name,
				"description": description,
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/tags", ws), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "Tag slug (required).")
	c.Flags().StringVar(&description, "description", "", "Human description (required, 6–280 chars).")
	_ = c.MarkFlagRequired("name")
	_ = c.MarkFlagRequired("description")
	return c
}

func editTagCmd() *cobra.Command {
	var newName, description string

	c := &cobra.Command{
		Use:   "edit-tag <tag-slug>",
		Short: "Rename a tag and/or update its description.",
		Long: `Pass --name to rename the tag (all attached docs are migrated
server-side). Pass --description to overwrite the description.
At least one must be provided.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if newName == "" && !cmd.Flags().Changed("description") {
				return fmt.Errorf("provide --name and/or --description")
			}
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
				body["new_slug"] = newName
			}
			if cmd.Flags().Changed("description") {
				body["description"] = description
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Patch(ctx,
				fmt.Sprintf("/api/workspaces/%s/tags/%s", ws, args[0]), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&newName, "name", "", "New tag slug.")
	c.Flags().StringVar(&description, "description", "", "New description (pass empty string to clear).")
	return c
}

func deleteTagCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "delete-tag <tag-slug>",
		Short:        "Delete a tag (detaches it from all docs).",
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
				fmt.Sprintf("/api/workspaces/%s/tags/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}
