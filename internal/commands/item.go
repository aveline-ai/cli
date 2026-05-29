package commands

import (
	"fmt"

	"github.com/aveline-ai/cli/internal/client"
	"github.com/aveline-ai/cli/internal/output"
	"github.com/aveline-ai/cli/internal/slug"
	"github.com/spf13/cobra"
)

func newListCmd(g *Globals) *cobra.Command {
	var (
		pinned bool
		tags   []string
		view   string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List items in the active workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			items, err := c.ListItems(cmd.Context(), ws, client.ListItemsParams{
				Pinned: pinned, Tags: tags, View: view,
			})
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), map[string]any{"items": items})
			}
			output.Items(cmd.OutOrStdout(), items)
			return nil
		},
	}
	cmd.Flags().BoolVar(&pinned, "pinned", false, "only pinned items")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "filter by tag (repeatable)")
	cmd.Flags().StringVar(&view, "view", "", "filter by view slug")
	return cmd
}

func newGetCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "get <slug>",
		Short: "Print an item",
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
			it, err := c.GetItem(cmd.Context(), ws, args[0])
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), it)
			}
			output.Item(cmd.OutOrStdout(), it)
			return nil
		},
	}
}

func newSaveCmd(g *Globals) *cobra.Command {
	var (
		title   string
		body    string
		tags    []string
		pin     bool
		summary string
		slugIn  string
	)
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Create a new item",
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}

			bodyStr, _, err := readBody(body, cmd.InOrStdin())
			if err != nil {
				return err
			}

			itemSlug := slugIn
			if itemSlug == "" {
				itemSlug, err = slug.From(title)
				if err != nil {
					return err
				}
			}

			req := client.CreateItemRequest{
				Title:  title,
				Body:   bodyStr,
				Tags:   tags,
				Pinned: pin,
				Slug:   itemSlug,
			}
			if cmd.Flags().Changed("summary") {
				s := summary
				req.Summary = &s
			}

			c, err := g.Client()
			if err != nil {
				return err
			}
			it, err := c.CreateItem(cmd.Context(), ws, req)
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), it)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "saved: %s\n", it.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "item title (required)")
	cmd.Flags().StringVar(&body, "body", "", "body source: '-' for stdin, or a file path")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tag (repeatable)")
	cmd.Flags().BoolVar(&pin, "pin", false, "pin the item")
	cmd.Flags().StringVar(&summary, "summary", "", "one-line summary")
	cmd.Flags().StringVar(&slugIn, "slug", "", "explicit slug (default: derived from title)")
	return cmd
}

func newEditCmd(g *Globals) *cobra.Command {
	var (
		title      string
		body       string
		addTags    []string
		removeTags []string
		pin        bool
		unpin      bool
		summary    string
	)
	cmd := &cobra.Command{
		Use:   "edit <slug>",
		Short: "Edit an item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if pin && unpin {
				return fmt.Errorf("--pin and --unpin are mutually exclusive")
			}
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}

			req := client.UpdateItemRequest{}

			if cmd.Flags().Changed("title") {
				t := title
				req.Title = &t
			}

			if cmd.Flags().Changed("body") {
				bodyStr, _, err := readBody(body, cmd.InOrStdin())
				if err != nil {
					return err
				}
				req.Body = &bodyStr
			}

			if cmd.Flags().Changed("summary") {
				s := summary
				req.Summary = &s
			}

			if pin {
				v := true
				req.Pinned = &v
			} else if unpin {
				v := false
				req.Pinned = &v
			}

			if len(addTags) > 0 || len(removeTags) > 0 {
				cur, err := c.GetItem(cmd.Context(), ws, args[0])
				if err != nil {
					return err
				}
				newTags := mergeTags(cur.Tags, addTags, removeTags)
				req.Tags = &newTags
			}

			it, err := c.UpdateItem(cmd.Context(), ws, args[0], req)
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), it)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated: %s\n", it.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&body, "body", "", "body source: '-' for stdin, or a file path")
	cmd.Flags().StringSliceVar(&addTags, "add-tag", nil, "tag to add (repeatable)")
	cmd.Flags().StringSliceVar(&removeTags, "remove-tag", nil, "tag to remove (repeatable)")
	cmd.Flags().BoolVar(&pin, "pin", false, "pin the item")
	cmd.Flags().BoolVar(&unpin, "unpin", false, "unpin the item")
	cmd.Flags().StringVar(&summary, "summary", "", "set summary")
	return cmd
}

func newDeleteCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <slug>",
		Short: "Soft-delete an item",
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
			if err := c.DeleteItem(cmd.Context(), ws, args[0]); err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), map[string]any{"slug": args[0], "deleted": true})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted: %s\n", args[0])
			return nil
		},
	}
}

func newRestoreCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "restore <slug>",
		Short: "Restore a soft-deleted item",
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
			it, err := c.RestoreItem(cmd.Context(), ws, args[0])
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), it)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "restored: %s\n", it.Slug)
			return nil
		},
	}
}

// mergeTags applies add-tag/remove-tag to an existing tag list, preserving
// the original ordering for tags that survive, then appending net-new tags
// in --add-tag order. Duplicates are deduped (set semantics).
func mergeTags(current, add, remove []string) []string {
	removeSet := make(map[string]struct{}, len(remove))
	for _, t := range remove {
		removeSet[t] = struct{}{}
	}
	seen := make(map[string]struct{}, len(current)+len(add))
	out := make([]string, 0, len(current)+len(add))
	for _, t := range current {
		if _, drop := removeSet[t]; drop {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	for _, t := range add {
		if _, drop := removeSet[t]; drop {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
