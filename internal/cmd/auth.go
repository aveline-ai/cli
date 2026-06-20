package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aveline-ai/cli/internal/api"
	"github.com/aveline-ai/cli/internal/config"
	"github.com/spf13/cobra"
)

func loginCmd() *cobra.Command {
	var token string
	var apiURL string

	c := &cobra.Command{
		Use:   "login",
		Short: "Save an API token + API URL to the config file.",
		Long: `Persist your API token (and optionally API URL) so future verbs
authenticate automatically.

Reads from --token, $AVELINE_TOKEN, or stdin (interactive). Stores
the token in ~/.config/aveline/config.toml with mode 0600.

After login, run ` + "`aveline whoami`" + ` to verify, then
` + "`aveline use-workspace <slug>`" + ` to pick a default workspace.`,
		Example: `  aveline login --token avl_abc...
  echo $AVELINE_TOKEN | aveline login
  aveline login --api-url https://aveline.staging.example.com --token avl_...`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadCfg()

			// Token precedence: flag > env > stdin
			if token == "" {
				if env := os.Getenv("AVELINE_TOKEN"); env != "" {
					token = env
				} else {
					fmt.Fprint(os.Stderr, "API token: ")
					s := bufio.NewScanner(os.Stdin)
					if s.Scan() {
						token = strings.TrimSpace(s.Text())
					}
				}
			}
			if token == "" {
				return fmt.Errorf("no token provided")
			}
			if !strings.HasPrefix(token, "avl_") {
				return fmt.Errorf("tokens start with 'avl_' — got %q", truncate(token, 12))
			}

			cfg.Token = token
			if apiURL != "" {
				cfg.APIURL = strings.TrimRight(apiURL, "/")
			}
			if err := config.Save(cfg); err != nil {
				return err
			}

			// Quick whoami to confirm + give the agent feedback.
			ctx, cancel := withTimeout()
			defer cancel()
			client := api.New(cfg.Resolve(flagAPIURL), token)
			raw, apiErr := client.Get(ctx, "/api/me", nil)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&token, "token", "", "API token (avl_...). If omitted, read from $AVELINE_TOKEN or stdin.")
	c.Flags().StringVar(&apiURL, "api-url", "", "API URL to persist in the config (overrides the existing config value).")
	return c
}

func logoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "logout",
		Short:        "Clear the saved token (workspace + API URL preserved).",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadCfg()
			cfg.Token = ""
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, `{"ok":true}`)
			return nil
		},
	}
}

func whoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "whoami",
		Short:        "GET /api/me — current user + accessible workspaces.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, "/api/me", nil)
			return handle(raw, apiErr)
		},
	}
}

func heartbeatCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "heartbeat",
		Short:        "GET /api/heartbeat — service + version (no auth).",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(true)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, "/api/heartbeat", nil)
			return handle(raw, apiErr)
		},
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
