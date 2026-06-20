package cmd

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

func listEventsCmd() *cobra.Command {
	var (
		limit    int
		beforeID string
	)

	c := &cobra.Command{
		Use:   "list-events",
		Short: "List activity events in the current workspace (newest first).",
		Long: `Stream workspace activity (doc-created / doc-version-shipped /
comment-posted / etc.). Page through history with --before-id (use
the last event id from the previous page).`,
		Example:      `  aveline list-events --limit 50`,
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
			q := url.Values{}
			if cmd.Flags().Changed("limit") {
				q.Set("limit", strconv.Itoa(limit))
			}
			if beforeID != "" {
				q.Set("before_id", beforeID)
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx,
				fmt.Sprintf("/api/workspaces/%s/events", ws), q)
			return handle(raw, apiErr)
		},
	}

	c.Flags().IntVar(&limit, "limit", 50, "Max events to return (1-200).")
	c.Flags().StringVar(&beforeID, "before-id", "", "Paginate: return events with id < this.")
	return c
}
