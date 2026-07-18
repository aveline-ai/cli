package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func listKeysCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-keys",
		Short:        "List your API keys (masked).",
		Example:      `  aveline list-keys`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, "/api/keys", nil)
			return handle(raw, apiErr)
		},
	}
}

func createKeyCmd() *cobra.Command {
	var name string

	c := &cobra.Command{
		Use:   "create-key",
		Short: "Mint a new API key; the plaintext is shown exactly once.",
		Long: `Mint a new API key. The response's "key" field is the plaintext —
shown exactly once, only its hash persists. Use this to rotate your own
credential: create-key, switch to the new key, then revoke-key the old
one (revoking the last active key is refused: last_key).`,
		Example:      `  aveline create-key --name laptop`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx, "/api/keys", map[string]any{"name": name})
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&name, "name", "", "Key name, e.g. \"laptop\" or \"ci\". Required.")
	_ = c.MarkFlagRequired("name")
	return c
}

func revokeKeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke-key <key-id>",
		Short: "Revoke one of your API keys by id (see list-keys).",
		Long: `Revoke one of your API keys by id. Anything still using it stops
authenticating immediately. The last active key can't be revoked
(last_key) — mint a replacement first.`,
		Example:      `  aveline revoke-key 6e5f6f0e-...`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Delete(ctx, fmt.Sprintf("/api/keys/%s", args[0]))
			return handle(raw, apiErr)
		},
	}
}
