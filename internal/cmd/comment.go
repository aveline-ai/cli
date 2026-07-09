package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func listCommentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list-comments <doc-slug>",
		Short:        "List comments on a doc.",
		Args:         cobra.ExactArgs(1),
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
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx,
				fmt.Sprintf("/api/workspaces/%s/docs/%s/comments", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func createCommentCmd() *cobra.Command {
	var (
		body, blockID, actor string
	)

	c := &cobra.Command{
		Use:   "create-comment <doc-slug>",
		Short: "Post a new top-level comment on a doc.",
		Long: `Post a new top-level comment. Omit --block-id for a doc-level
comment; pass --block-id b_xxx to anchor it to a specific block (id
comes from get-doc).

--body accepts text directly, '-' for stdin, or @PATH for a file.

Returns {"ok": true, "id": "<base_comment_id>"} — that's the stable
id all other comment verbs reference.`,
		Example: `  aveline create-comment deploy-guide --body "Worth adding a rollback note?"
  aveline create-comment deploy-guide --block-id b_abc --body "Inline question"`,
		Args:         cobra.ExactArgs(1),
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
			text, err := readTextInput(body)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"body":  text,
				"actor": defaultStr(actor, "agent"),
			}
			if blockID != "" {
				payload["block_id"] = blockID
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/docs/%s/comments", ws, args[0]), payload)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&body, "body", "", "Comment text. '-' = stdin, @PATH = file.")
	c.Flags().StringVar(&blockID, "block-id", "", "Anchor to a block (omit for doc-level).")
	c.Flags().StringVar(&actor, "actor", "agent", "Actor type: human | agent.")
	_ = c.MarkFlagRequired("body")
	return c
}

func replyCommentCmd() *cobra.Command {
	var (
		body, actor string
	)

	c := &cobra.Command{
		Use:   "reply-comment <doc-slug> <parent-comment-id>",
		Short: "Reply to a comment by its id.",
		Long: `Reply to an existing comment by passing its id (from list-comments
or the create-comment response). The reply inherits the parent's
block anchor automatically.`,
		Args:         cobra.ExactArgs(2),
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
			text, err := readTextInput(body)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"body":              text,
				"parent_comment_id": args[1],
				"actor":             defaultStr(actor, "agent"),
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/docs/%s/comments", ws, args[0]), payload)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&body, "body", "", "Reply text. '-' = stdin, @PATH = file.")
	c.Flags().StringVar(&actor, "actor", "agent", "Actor type: human | agent.")
	_ = c.MarkFlagRequired("body")
	return c
}

func editCommentCmd() *cobra.Command {
	var body string

	c := &cobra.Command{
		Use:          "edit-comment <comment-id>",
		Short:        "Rewrite a comment's body (author-only).",
		Args:         cobra.ExactArgs(1),
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
			text, err := readTextInput(body)
			if err != nil {
				return err
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Patch(ctx,
				fmt.Sprintf("/api/workspaces/%s/comments/%s", ws, args[0]),
				map[string]any{"body": text})
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&body, "body", "", "New body. '-' = stdin, @PATH = file.")
	_ = c.MarkFlagRequired("body")
	return c
}

func deleteCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "delete-comment <comment-id>",
		Short:        "Soft-delete your own comment.",
		Args:         cobra.ExactArgs(1),
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
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Delete(ctx,
				fmt.Sprintf("/api/workspaces/%s/comments/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}

func undeleteCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "undelete-comment <comment-id>",
		Short:        "Restore a previously-deleted comment.",
		Args:         cobra.ExactArgs(1),
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
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/comments/%s/undelete", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func resolveCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve-comment <comment-id>",
		Short: "Mark a comment thread resolved.",
		Long: `Mark resolved. For agent-driven resolution that ships ALONG WITH
a doc-version change, prefer the disposition flow in
` + "`aveline edit-doc`" + ` — that posts a reply comment in the same
transaction and pins it to the new doc-version. This verb is the
standalone "I addressed this offline" path.`,
		Args:         cobra.ExactArgs(1),
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
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/comments/%s/resolve", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func unresolveCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "unresolve-comment <comment-id>",
		Short:        "Re-open a resolved thread.",
		Args:         cobra.ExactArgs(1),
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
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/comments/%s/unresolve", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

// ----- Shared helper --------------------------------------------------

// readTextInput accepts a flag value that may be:
//
//   - "-": read from stdin
//   - "@PATH": read the file at PATH
//   - anything else: the literal text
func readTextInput(arg string) (string, error) {
	switch {
	case arg == "":
		return "", fmt.Errorf("empty input")
	case arg == "-":
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case len(arg) > 1 && arg[0] == '@':
		b, err := os.ReadFile(arg[1:])
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return arg, nil
	}
}
