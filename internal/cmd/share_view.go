package cmd

// View permissions: the doc share model copied onto views. viewer can
// use the view; editor can also edit its config. Owner-only mutations;
// the API enforces everything, these are thin.

import (
	"fmt"

	"github.com/spf13/cobra"
)

func setViewVisibilityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-view-visibility <name> <private|workspace>",
		Short: "Change who can see a view: just you (private) or the whole team (workspace).",
		Long: `Change a view's visibility in place. Owner only (the owner is the
view's original creator); no new version is created.

  private    only you, plus anyone you share-view with
  workspace  everyone in the workspace (the default)

A pinned view can't go private (the sidebar is a team surface): unpin
it first.`,
		Example: `  aveline set-view-visibility my-daily private
  aveline set-view-visibility my-daily workspace`,
		Args:         cobra.ExactArgs(2),
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
			raw, apiErr := client.Put(ctx,
				fmt.Sprintf("/api/workspaces/%s/views/%s/visibility", ws, args[0]),
				map[string]any{"visibility": args[1]})
			return handle(raw, apiErr)
		},
	}
}

func shareViewCmd() *cobra.Command {
	var role string

	c := &cobra.Command{
		Use:   "share-view <name>",
		Short: "Grant a workspace member access to one of your private views.",
		Long: `Grant a specific member access to a view you own. Meaningful on
private views (workspace views are already visible to everyone).
Sharing again with a different --role updates the existing grant.

  viewer  can use the view
  editor  can also edit its config`,
		Example: `  aveline share-view my-daily --user trevor
  aveline share-view my-daily --user kyle --role editor`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := cmd.Flags().GetString("user")
			if err != nil || user == "" {
				return fmt.Errorf("--user <username> is required")
			}
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
				fmt.Sprintf("/api/workspaces/%s/views/%s/shares", ws, args[0]),
				map[string]any{"username": user, "role": role})
			return handle(raw, apiErr)
		},
	}

	c.Flags().String("user", "", "Username of the member to share with (required).")
	c.Flags().StringVar(&role, "role", "viewer", "viewer (read + comment) | editor (also edit).")
	return c
}

func unshareViewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:          "unshare-view <name>",
		Short:        "Revoke a member's access to one of your private views.",
		Example:      `  aveline unshare-view my-daily --user trevor`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := cmd.Flags().GetString("user")
			if err != nil || user == "" {
				return fmt.Errorf("--user <username> is required")
			}
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
				fmt.Sprintf("/api/workspaces/%s/views/%s/shares/%s", ws, args[0], user))
			return handle(raw, apiErr)
		},
	}

	c.Flags().String("user", "", "Username whose access to revoke (required).")
	return c
}

func listViewSharesCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-view-shares <name>",
		Short:        "List who a view is shared with (and its visibility).",
		Example:      `  aveline list-shares my-daily`,
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
				fmt.Sprintf("/api/workspaces/%s/views/%s/shares", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}
