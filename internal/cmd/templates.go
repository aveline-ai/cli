package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

// templatesCmd lists the templates a workspace can start from, or shows
// exactly what one of them would create, so the choice can be made by
// someone who has seen the answer.
func templatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates [<name>]",
		Short: "List the workspace templates, or show exactly what one would create.",
		Long: `List the templates a new workspace can start from, or, with <name>,
show exactly what that template would create: its tags, its views, its
docs and the orientation doc it writes.

The templates are "autonomous-dev" (recommended), "knowledge-base" and
"blank". Pass the one you choose to create-workspace --template.

Reads nothing into any workspace and changes nothing. Workspace-
independent (it describes the templates, not any workspace's data), so
no -w is needed.`,
		Example: `  aveline templates | jq '.templates[] | {name, recommended}'
  aveline templates autonomous-dev
  aveline templates blank | jq .template.orientation_blocks`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveAPI(false)
			if err != nil {
				return err
			}
			path := "/api/templates"
			if len(args) == 1 {
				path += "/" + url.PathEscape(args[0])
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, path, nil)
			return handle(raw, apiErr)
		},
	}
}
