package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func listMembersCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-members",
		Short:        "List members of the current workspace.",
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
				fmt.Sprintf("/api/workspaces/%s/members", ws), nil)
			return handle(raw, apiErr)
		},
	}
}

func addMemberCmd() *cobra.Command {
	var username string

	c := &cobra.Command{
		Use:          "add-member",
		Short:        "Add an existing user to the current workspace by username.",
		Example:      `  aveline add-member --username alice`,
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
				fmt.Sprintf("/api/workspaces/%s/members", ws),
				map[string]any{"username": username})
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&username, "username", "", "Username of an existing user (required).")
	_ = c.MarkFlagRequired("username")
	return c
}

func removeMemberCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "remove-member <username>",
		Short:        "Remove a member from the current workspace.",
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
				fmt.Sprintf("/api/workspaces/%s/members/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}

func getInviteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-invite",
		Short: "Get (or generate) the workspace invite URL.",
		Long: `Returns the current workspace invite — code + shareable URL. If
no active invite exists, one is created. Idempotent.`,
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
				fmt.Sprintf("/api/workspaces/%s/invite", ws), nil)
			return handle(raw, apiErr)
		},
	}
}

func revokeInviteCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "revoke-invite",
		Short:        "Revoke the current workspace invite code.",
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
				fmt.Sprintf("/api/workspaces/%s/invite", ws))
			return handle(raw, apiErr)
		},
	}
}
