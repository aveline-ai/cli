package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func listVersionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-versions <doc-slug>",
		Short:        "List all versions of a doc (newest first).",
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s/versions", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func getVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-version <doc-slug> <version-number>",
		Short: "Show a specific version of a doc (full block body).",
		Long: `Pinpoint a historical state. Pair with list-versions to find the
version number you want.`,
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
			raw, apiErr := client.Get(ctx,
				fmt.Sprintf("/api/workspaces/%s/docs/%s/versions/%s", ws, args[0], args[1]), nil)
			return handle(raw, apiErr)
		},
	}
}
