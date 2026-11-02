package cmd

import (
	"github.com/spf13/cobra"
)

// guideCmd prints Aveline's own guide to using Aveline. It is versioned
// and opinion-free: it says what Aveline is and how its pieces fit, not
// how any team should work.
func guideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "Print Aveline's own guide to using Aveline (versioned, opinion-free).",
		Long: `Print Aveline's own guide to using Aveline: what docs, tags, views and
comments are and how they fit together. It is versioned (the response
carries "version" and a "changelog") and opinion-free: it describes the
product, not how a team should work.

Read it on your first session in a workspace (get-orientation reports
"first_session": true), or when the person asks how Aveline works. The
workspace's orientation doc is the authority on how THIS team works; when
the two differ, the team's doc wins.

Workspace-independent (it describes Aveline, not any workspace's data),
so no -w is needed.`,
		Example: `  aveline guide
  aveline guide | jq '.guide.version, .guide.changelog'`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, "/api/guide", nil)
			return handle(raw, apiErr)
		},
	}
}
