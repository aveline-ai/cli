package cmd

// Doc permissions v1: visibility (private | workspace) plus per-member
// shares. "Shared with some people" is a private doc with share rows.
// Owner-only mutations; the API enforces everything, these are thin.

import (
	"fmt"

	"github.com/spf13/cobra"
)

func setDocVisibilityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-doc-visibility <slug> <private|workspace>",
		Short: "Change who can see a doc: just you (private) or the whole team (workspace).",
		Long: `Change a doc's visibility in place. Owner only; no new version is
created (visibility is placement-style state, like pin slots).

  private    only you, plus anyone you share-doc with
  workspace  everyone in the workspace (the default for new docs)

A pinned doc can't go private (the home page is a team surface): unpin
first. The orientation doc is always workspace-visible.`,
		Example: `  aveline set-doc-visibility my-planning-notes private
  aveline set-doc-visibility my-planning-notes workspace`,
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s/visibility", ws, args[0]),
				map[string]any{"visibility": args[1]})
			return handle(raw, apiErr)
		},
	}
}

func shareDocCmd() *cobra.Command {
	var role string

	c := &cobra.Command{
		Use:   "share-doc <slug>",
		Short: "Grant a workspace member access to one of your private docs.",
		Long: `Grant a specific member access to a doc you own. Meaningful on
private docs (workspace docs are already visible to everyone).
Sharing again with a different --role updates the existing grant.

  viewer  reads and comments
  editor  reads, comments, and edits`,
		Example: `  aveline share-doc my-planning-notes --user trevor
  aveline share-doc my-planning-notes --user kyle --role editor`,
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s/shares", ws, args[0]),
				map[string]any{"username": user, "role": role})
			return handle(raw, apiErr)
		},
	}

	c.Flags().String("user", "", "Username of the member to share with (required).")
	c.Flags().StringVar(&role, "role", "viewer", "viewer (read + comment) | editor (also edit).")
	return c
}

func unshareDocCmd() *cobra.Command {
	c := &cobra.Command{
		Use:          "unshare-doc <slug>",
		Short:        "Revoke a member's access to one of your private docs.",
		Example:      `  aveline unshare-doc my-planning-notes --user trevor`,
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s/shares/%s", ws, args[0], user))
			return handle(raw, apiErr)
		},
	}

	c.Flags().String("user", "", "Username whose access to revoke (required).")
	return c
}

func listSharesCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-shares <slug>",
		Short:        "List who a doc is shared with (and its visibility).",
		Example:      `  aveline list-shares my-planning-notes`,
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s/shares", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}
