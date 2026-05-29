package commands

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aveline-ai/cli/internal/client"
	"github.com/aveline-ai/cli/internal/config"
	"github.com/spf13/cobra"
)

func newLoginCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Store an API token in the config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := g.Config()
			if err != nil {
				return err
			}
			url := cfg.Resolve(g.APIURL)
			fmt.Fprintf(cmd.OutOrStdout(), "Aveline API: %s\n", url)
			fmt.Fprint(cmd.OutOrStdout(), "Paste API token (avl_...): ")

			scanner := bufio.NewScanner(cmd.InOrStdin())
			if !scanner.Scan() {
				return fmt.Errorf("no token provided")
			}
			token := strings.TrimSpace(scanner.Text())
			if token == "" {
				return fmt.Errorf("no token provided")
			}

			// Sanity check: hit /api/me with the new token.
			c := client.New(url, token)
			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()
			me, err := c.Me(ctx)
			if err != nil {
				return fmt.Errorf("verifying token: %w", err)
			}

			cfg.APIURL = url
			cfg.Token = token
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "logged in as %s <%s>\n", me.User.Username, me.User.Email)
			return nil
		},
	}
}
