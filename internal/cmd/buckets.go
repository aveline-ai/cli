package cmd

// View buckets (see the view-buckets TIP): the space a view lives in
// and the unit views are shared at. Team = everyone (the default),
// yours = just you, project buckets = binary membership managed by
// their owner.

import (
	"fmt"

	"github.com/spf13/cobra"
)

func listBucketsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-buckets",
		Short: "List the view buckets you can use (team, yours, your projects).",
		Long: `List every bucket you can use, with kind, owner, and (for project
buckets) the member list. Buckets are where views live: team is
everyone, your personal bucket is just you, and project buckets share
all their views with their members.`,
		Example:      `  aveline list-buckets`,
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
			raw, apiErr := client.Get(ctx, fmt.Sprintf("/api/workspaces/%s/view-buckets", ws), nil)
			return handle(raw, apiErr)
		},
	}
}

func createBucketCmd() *cobra.Command {
	var visibility string

	c := &cobra.Command{
		Use:   "create-bucket <name>",
		Short: "Create a project bucket: a shared space for views.",
		Long: `Create a project bucket. Add members with add-bucket-member; they
can use (and edit) every view in the bucket, present and future. Move
views in with move-view or create them there with
create-view --bucket. Names "team" and "personal-*" are reserved.

Visibility uses the same enum docs do:

  private    owner + invited members (the default)
  workspace  everyone in the workspace, current and future`,
		Example: `  aveline create-bucket launch
  aveline create-bucket launch --visibility workspace
  aveline create-view --name launch-tickets --description "Launch work by status." --tag launch --bucket launch`,
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
			body := map[string]any{"name": args[0]}
			if visibility != "" {
				body["visibility"] = visibility
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/view-buckets", ws), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&visibility, "visibility", "", "private (owner + members, default) | workspace (everyone).")
	return c
}

func setBucketVisibilityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-bucket-visibility <name> <private|workspace>",
		Short: "Change who can see a project bucket: its members or the whole team.",
		Long: `Change a project bucket's visibility in place. Owner only.

  private    owner + invited members
  workspace  everyone in the workspace, current and future

Team is always workspace-visible and personal buckets are always
private; only project buckets change.`,
		Example: `  aveline set-bucket-visibility launch workspace
  aveline set-bucket-visibility launch private`,
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
				fmt.Sprintf("/api/workspaces/%s/view-buckets/%s/visibility", ws, args[0]),
				map[string]any{"visibility": args[1]})
			return handle(raw, apiErr)
		},
	}
}

func deleteBucketCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "delete-bucket <name>",
		Short:        "Delete an empty project bucket (owner only).",
		Long:         `Delete a project bucket. Owner only, and it must hold no views: move or delete them first. Team and personal buckets can't be deleted.`,
		Example:      `  aveline delete-bucket launch`,
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
				fmt.Sprintf("/api/workspaces/%s/view-buckets/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}

func addBucketMemberCmd() *cobra.Command {
	c := &cobra.Command{
		Use:          "add-bucket-member <bucket>",
		Short:        "Add a workspace member to a project bucket (owner only).",
		Long:         `Add a member to a project bucket. Binary membership: being in the bucket means using and editing every view it holds, now and later.`,
		Example:      `  aveline add-bucket-member launch --user trevor`,
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
				fmt.Sprintf("/api/workspaces/%s/view-buckets/%s/members", ws, args[0]),
				map[string]any{"username": user})
			return handle(raw, apiErr)
		},
	}

	c.Flags().String("user", "", "Username of the member to add (required).")
	return c
}

func removeBucketMemberCmd() *cobra.Command {
	c := &cobra.Command{
		Use:          "remove-bucket-member <bucket>",
		Short:        "Remove a member from a project bucket (owner only).",
		Example:      `  aveline remove-bucket-member launch --user trevor`,
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
				fmt.Sprintf("/api/workspaces/%s/view-buckets/%s/members/%s", ws, args[0], user))
			return handle(raw, apiErr)
		},
	}

	c.Flags().String("user", "", "Username of the member to remove (required).")
	return c
}

func moveViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "move-view <view-name> <bucket>",
		Short: "Move a view to another bucket (view owner only).",
		Long: `Move a view into another bucket, in place (no version minted). The
view's owner only, and only into a bucket they can use. "yours"
targets your personal bucket, creating it if needed.`,
		Example: `  aveline move-view my-daily yours
  aveline move-view launch-tickets launch
  aveline move-view launch-tickets team`,
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
				fmt.Sprintf("/api/workspaces/%s/views/%s/bucket", ws, args[0]),
				map[string]any{"bucket": args[1]})
			return handle(raw, apiErr)
		},
	}
}
