package commands

import (
	"fmt"

	"github.com/aveline-ai/cli/internal/client"
	"github.com/aveline-ai/cli/internal/output"
	"github.com/spf13/cobra"
)

func newReplyCmd(g *Globals) *cobra.Command {
	var bodyFlag string

	cmd := &cobra.Command{
		Use:   "reply <item-slug>",
		Short: "Post a reply on an item's thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, _, err := readBody(bodyFlag, cmd.InOrStdin())
			if err != nil {
				return err
			}
			if body == "" {
				return fmt.Errorf("--body is required (pass `-` to read stdin or a path)")
			}

			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}

			msg, err := c.CreateMessage(cmd.Context(), ws, args[0], client.CreateMessageRequest{Body: body})
			if err != nil {
				return err
			}

			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), msg)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "posted reply by %s\n", msg.Author.Username)
			return nil
		},
	}
	cmd.Flags().StringVar(&bodyFlag, "body", "", "reply body (- for stdin, or a file path)")
	return cmd
}

func newThreadCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "thread <item-slug>",
		Short: "Show all replies on an item",
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
			msgs, err := c.ListMessages(cmd.Context(), ws, args[0])
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), map[string]any{"messages": msgs})
			}
			output.Thread(cmd.OutOrStdout(), msgs)
			return nil
		},
	}
}

func newDeleteReplyCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "delete-reply <item-slug> <id>",
		Short: "Soft-delete a reply",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			if err := c.DeleteMessage(cmd.Context(), ws, args[0], args[1]); err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), map[string]any{"id": args[1], "deleted": true})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted reply %s\n", args[1])
			return nil
		},
	}
}

func newEditReplyCmd(g *Globals) *cobra.Command {
	var bodyFlag string

	cmd := &cobra.Command{
		Use:   "edit-reply <item-slug> <id>",
		Short: "Edit a reply (--body required)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, _, err := readBody(bodyFlag, cmd.InOrStdin())
			if err != nil {
				return err
			}
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			ws, err := g.WorkspaceSlug()
			if err != nil {
				return err
			}
			c, err := g.Client()
			if err != nil {
				return err
			}
			msg, err := c.UpdateMessage(cmd.Context(), ws, args[0], args[1], client.UpdateMessageRequest{Body: body})
			if err != nil {
				return err
			}
			if g.JSONOut {
				return output.JSON(cmd.OutOrStdout(), msg)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "edited reply %s\n", msg.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&bodyFlag, "body", "", "new reply body (- for stdin, or a file path)")
	return cmd
}
