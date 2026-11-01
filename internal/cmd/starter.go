package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

// starterCmd shows what `create-workspace --setup standard` would write,
// without writing it, so the choice can be made by someone who has seen
// the answer.
func starterCmd() *cobra.Command {
	var setup string

	c := &cobra.Command{
		Use:   "starter",
		Short: "Print what a new workspace would be set up with, without creating one.",
		Long: `Print the setup a new workspace would start with: the tags, the boards,
and the orientation doc that create-workspace would write.

--setup standard (the default) is the furnished version: a loop that
works on day one, the tags it refers to, and three boards. --setup
custom is the same orientation frame with nothing filled in.

Reads nothing and changes nothing. Workspace-independent, so no -w is
needed.`,
		Example: `  aveline starter | jq '.starter.tags[].slug'
  aveline starter --setup custom`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			query := url.Values{}
			if setup != "" {
				query.Set("setup", setup)
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, "/api/starter", query)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&setup, "setup", "", `Which setup to show: "standard" (default) or "custom".`)
	return c
}
