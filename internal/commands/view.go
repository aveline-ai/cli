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

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List saved views",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			views, err := c.ListViews(cmd.Context(), ws)
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), map[string]any{"views": views})
			}
			output.Views(cmd.OutOrStdout(), views)
			return nil
		},
	})

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
	)
	cmd := &cobra.Command{
		Use:   "create <slug>",
		Short: "Create a saved view",
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
			v, err := c.CreateView(cmd.Context(), ws, req)
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), v)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created view: %s\n", v.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "human label (required)")
	cmd.Flags().StringSliceVar(&tagFilter, "tag", nil, "tag in the filter (repeatable)")
	cmd.Flags().StringVar(&description, "description", "", "description")
	return cmd
}

func newViewEditCmd(g *Globals) *cobra.Command {
	var (
		name       string
		addTags    []string
		removeTags []string
	)
	cmd := &cobra.Command{
		Use:   "edit <slug>",
		Short: "Edit a saved view",
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
			v, err := c.UpdateView(cmd.Context(), ws, args[0], req)
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), v)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated view: %s\n", v.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().StringSliceVar(&addTags, "add-tag", nil, "tag to add (repeatable)")
	cmd.Flags().StringSliceVar(&removeTags, "remove-tag", nil, "tag to remove (repeatable)")
	return cmd
}
