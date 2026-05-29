package commands

import (
	"fmt"

	"github.com/aveline-ai/cli/internal/client"
	"github.com/aveline-ai/cli/internal/output"
	"github.com/spf13/cobra"
)

func newViewCmd(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Manage saved views",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List saved views (yours + team, by default)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			scope := ""
			if mineOnly, _ := cmd.Flags().GetBool("mine"); mineOnly {
				scope = "personal"
			}
			if teamOnly, _ := cmd.Flags().GetBool("team"); teamOnly {
				scope = "team"
			}
			views, err := c.ListViews(cmd.Context(), ws, client.ListViewsParams{Scope: scope})
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), map[string]any{"views": views})
			}
			output.Views(cmd.OutOrStdout(), views)
			return nil
		},
	}
	listCmd.Flags().Bool("mine", false, "only personal views you created")
	listCmd.Flags().Bool("team", false, "only team views")
	cmd.AddCommand(listCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "get <slug>",
		Short: "Show a view and the items it matches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			out, err := c.GetView(cmd.Context(), ws, args[0])
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), out)
			}
			output.ViewWithItems(cmd.OutOrStdout(), out)
			return nil
		},
	})

	cmd.AddCommand(newViewCreateCmd(g))
	cmd.AddCommand(newViewEditCmd(g))

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <slug>",
		Short: "Soft-delete a view",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			if err := c.DeleteView(cmd.Context(), ws, args[0]); err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), map[string]any{"slug": args[0], "deleted": true})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted view: %s\n", args[0])
			return nil
		},
	})

	return cmd
}

func newViewCreateCmd(g *Globals) *cobra.Command {
	var (
		name        string
		tagFilter   []string
		description string
		share       bool
	)
	cmd := &cobra.Command{
		Use:   "create <slug>",
		Short: "Create a saved view (personal by default; --share to make it a team view)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			req := client.CreateViewRequest{
				Slug:      args[0],
				Name:      name,
				TagFilter: tagFilter,
			}
			if cmd.Flags().Changed("description") {
				d := description
				req.Description = &d
			}
			if share {
				req.Scope = "team"
			} else {
				req.Scope = "personal"
			}
			v, err := c.CreateView(cmd.Context(), ws, req)
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), v)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created view: %s (%s)\n", v.Slug, v.Scope)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "human label (required)")
	cmd.Flags().StringSliceVar(&tagFilter, "tag", nil, "tag in the filter (repeatable)")
	cmd.Flags().StringVar(&description, "description", "", "description")
	cmd.Flags().BoolVar(&share, "share", false, "make this a team view (visible to everyone in the workspace)")
	return cmd
}

func newViewEditCmd(g *Globals) *cobra.Command {
	var (
		name       string
		addTags    []string
		removeTags []string
		share      bool
		unshare    bool
	)
	cmd := &cobra.Command{
		Use:   "edit <slug>",
		Short: "Edit a saved view (--share to promote to team, --unshare to demote to personal)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if share && unshare {
				return fmt.Errorf("--share and --unshare are mutually exclusive")
			}
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			req := client.UpdateViewRequest{}
			if cmd.Flags().Changed("name") {
				n := name
				req.Name = &n
			}
			if len(addTags) > 0 || len(removeTags) > 0 {
				cur, err := c.GetView(cmd.Context(), ws, args[0])
				if err != nil {
					return err
				}
				merged := mergeTags(cur.View.TagFilter, addTags, removeTags)
				req.TagFilter = &merged
			}
			if share {
				s := "team"
				req.Scope = &s
			}
			if unshare {
				s := "personal"
				req.Scope = &s
			}
			v, err := c.UpdateView(cmd.Context(), ws, args[0], req)
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), v)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated view: %s (%s)\n", v.Slug, v.Scope)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().StringSliceVar(&addTags, "add-tag", nil, "tag to add (repeatable)")
	cmd.Flags().StringSliceVar(&removeTags, "remove-tag", nil, "tag to remove (repeatable)")
	cmd.Flags().BoolVar(&share, "share", false, "promote to team view (visible to all members)")
	cmd.Flags().BoolVar(&unshare, "unshare", false, "demote to personal view (only you)")
	return cmd
}
