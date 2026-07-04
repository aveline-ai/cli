package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/aveline-ai/cli/internal/api"
	"github.com/spf13/cobra"
)

func listDocsCmd() *cobra.Command {
	var (
		tags []string
		has  []string
	)

	c := &cobra.Command{
		Use:   "list-docs",
		Short: "List docs in the current workspace.",
		Long: `List docs in the current workspace.

--has filters by structural kind: "board" (the doc contains a board
block, i.e. renders a kanban).`,
		Example: `  aveline list-docs --tag runbook
  aveline list-docs --has board`,
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
			// Send as CSV. Phoenix's default URL-encoded parser
			// drops earlier values for repeated keys, so
			// `?tag=x&tag=y` becomes `tag=y` server-side. CSV
			// round-trips reliably.
			if len(tags) > 0 {
				q.Set("tag", strings.Join(tags, ","))
			}
			if len(has) > 0 {
				q.Set("has", strings.Join(has, ","))
			}

			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Get(ctx, fmt.Sprintf("/api/workspaces/%s/docs", ws), q)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringSliceVar(&tags, "tag", nil, "Filter by tag (repeatable).")
	c.Flags().StringSliceVar(&has, "has", nil, `Filter by kind: "board".`)
	return c
}

func getDocCmd() *cobra.Command {
	var follow bool

	c := &cobra.Command{
		Use:   "get-doc <slug>",
		Short: "Show a doc with full block body.",
		Long: `Show a doc with full block body.

--follow expands doc_link blocks: every linked doc's full body is
fetched, in chain order, and returned under "linked_docs". One level
deep only — links inside linked docs are not expanded. A stop whose
doc was deleted appears as {"deleted": true, ...} instead of a body.`,
		Example: `  aveline get-doc oncall-runbook
  aveline get-doc onboarding-story --follow`,
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s", ws, args[0]), nil)
			if apiErr != nil || !follow {
				return handle(raw, apiErr)
			}
			combined, apiErr := followDocLinks(ctx, client, ws, raw)
			return handle(combined, apiErr)
		},
	}

	c.Flags().BoolVar(&follow, "follow", false,
		"Also fetch every doc_link target's full body, in order (one level deep).")
	return c
}

func getOrientationCmd() *cobra.Command {
	var follow bool

	c := &cobra.Command{
		Use:   "get-orientation",
		Short: "Show the workspace orientation doc — how this workspace uses Aveline.",
		Long: `Show the workspace orientation doc: what lives in this workspace and
how the team works. Every workspace has one (well-known slug, seeded at
creation, undeletable).

If you're an agent starting a session in an unfamiliar workspace, run
this FIRST — with --follow to also pull the docs it links to, in order.`,
		Example:      `  aveline get-orientation --follow`,
		Args:         cobra.NoArgs,
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
				fmt.Sprintf("/api/workspaces/%s/orientation", ws), nil)
			if apiErr != nil || !follow {
				return handle(raw, apiErr)
			}
			combined, apiErr := followDocLinks(ctx, client, ws, raw)
			return handle(combined, apiErr)
		},
	}

	c.Flags().BoolVar(&follow, "follow", false,
		"Also fetch every doc_link target's full body, in order (one level deep).")
	return c
}

// followDocLinks walks the doc's doc_link blocks in order and fetches
// each target's full body, re-emitting the envelope with a
// "linked_docs" array appended. Targets the server echoed as deleted
// (and mid-chain 404 races) become {"deleted": true, ...} stubs so the
// chain never breaks.
func followDocLinks(ctx context.Context, client *api.Client, ws string, raw []byte) ([]byte, *api.Error) {
	var env struct {
		Doc map[string]any `json:"doc"`
	}
	if err := json.Unmarshal(raw, &env); err != nil || env.Doc == nil {
		return raw, nil // not the shape we expect; emit the plain doc
	}

	blocks, _ := env.Doc["blocks"].([]any)
	linked := []any{}

	for _, b := range blocks {
		bm, ok := b.(map[string]any)
		if !ok || bm["type"] != "doc_link" {
			continue
		}

		target, _ := bm["target"].(map[string]any)
		stub := map[string]any{"deleted": true, "doc_id": bm["doc_id"]}
		if target != nil {
			stub["slug"] = target["slug"]
			stub["title"] = target["title"]
		}

		if target == nil || target["deleted"] == true {
			linked = append(linked, stub)
			continue
		}

		slug, _ := target["slug"].(string)
		childRaw, apiErr := client.Get(ctx,
			fmt.Sprintf("/api/workspaces/%s/docs/%s", ws, slug), nil)
		if apiErr != nil {
			if apiErr.Code == "not_found" {
				linked = append(linked, stub)
				continue
			}
			return nil, apiErr
		}

		var childEnv struct {
			Doc json.RawMessage `json:"doc"`
		}
		if err := json.Unmarshal(childRaw, &childEnv); err != nil {
			continue
		}
		linked = append(linked, json.RawMessage(childEnv.Doc))
	}

	out, err := json.Marshal(map[string]any{
		"ok":          true,
		"doc":         env.Doc,
		"linked_docs": linked,
	})
	if err != nil {
		return nil, &api.Error{Code: "client_encode_error", Message: err.Error()}
	}
	return out, nil
}

func createDocCmd() *cobra.Command {
	var (
		title, slug, summary, intent, blocksPath, actor string
		tags                                            []string
	)

	c := &cobra.Command{
		Use:   "create-doc",
		Short: "Create a new doc.",
		Long: `Create a new doc. --blocks accepts either a JSON array on the
command line or a path to a file containing the JSON. Pass --blocks=-
to read from stdin.

Returns a minimal pointer the agent can chain off of:
    {"ok": true, "slug": "...", "doc_id": "...",
     "version_id": "...", "version_number": 1}`,
		Example: `  aveline create-doc \
    --title "Deploy guide" \
    --tag runbook --tag deploys \
    --blocks blocks.json \
    --intent "initial seed"`,
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
			blocks, err := readBlocks(blocksPath)
			if err != nil {
				return err
			}

			body := map[string]any{
				"title":  title,
				"blocks": blocks,
				"actor":  defaultStr(actor, "agent"),
			}
			if slug != "" {
				body["slug"] = slug
			}
			if summary != "" {
				body["summary"] = summary
			}
			if intent != "" {
				body["intent"] = intent
			}
			if len(tags) > 0 {
				body["tags"] = tags
			}

			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx, fmt.Sprintf("/api/workspaces/%s/docs", ws), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&title, "title", "", "Doc title (required).")
	c.Flags().StringVar(&slug, "slug", "", "URL slug (optional; derived from title if omitted).")
	c.Flags().StringVar(&summary, "summary", "", "One-line summary.")
	c.Flags().StringSliceVar(&tags, "tag", nil, "Tag (repeatable). Must exist in workspace.")
	c.Flags().StringVar(&blocksPath, "blocks", "", "Path to JSON file of blocks, '-' for stdin, or the raw JSON itself.")
	c.Flags().StringVar(&intent, "intent", "", "Why you're creating this doc (audit trail).")
	c.Flags().StringVar(&actor, "actor", "agent", "Actor type: human | agent.")
	_ = c.MarkFlagRequired("title")
	_ = c.MarkFlagRequired("blocks")
	return c
}

func applyOpsCmd() *cobra.Command {
	var (
		slug, opsPath, intent, dispositionsPath, actor string
		tags                                           []string
		title, summary                                 string
	)

	c := &cobra.Command{
		Use:   "apply-ops <slug>",
		Short: "Ship a new doc version by applying an ops array.",
		Long: `Apply a list of block operations to an existing doc, producing a
new version. Operations modify the block tree in order; supported ops:
append_block, insert_block, modify_block, delete_block, move_block.

--ops accepts either a JSON array on the command line, a path to a
file, or '-' for stdin.

--dispositions is required when ops touch a block carrying an open
comment. It accepts the same input forms as --ops. See
` + "`aveline list-comments`" + ` to inspect what's open.

Returns the new version pointer:
    {"ok": true, "slug": "...", "doc_id": "...",
     "version_id": "...", "version_number": N}`,
		Example: `  aveline apply-ops deploy-guide --ops ops.json \
    --intent "Add escalation section" \
    --dispositions disp.json`,
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
			slug = args[0]

			ops, err := readJSONInput(opsPath)
			if err != nil {
				return fmt.Errorf("reading --ops: %w", err)
			}

			body := map[string]any{
				"operations": ops,
				"actor":      defaultStr(actor, "agent"),
			}
			if intent != "" {
				body["intent"] = intent
			}
			if title != "" {
				body["title"] = title
			}
			if summary != "" {
				body["summary"] = summary
			}
			if len(tags) > 0 {
				body["tags"] = tags
			}
			if dispositionsPath != "" {
				disp, err := readJSONInput(dispositionsPath)
				if err != nil {
					return fmt.Errorf("reading --dispositions: %w", err)
				}
				body["comment_dispositions"] = disp
			}

			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Patch(ctx,
				fmt.Sprintf("/api/workspaces/%s/docs/%s", ws, slug), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().StringVar(&opsPath, "ops", "", "Ops array — path to JSON, '-' for stdin, or raw JSON.")
	c.Flags().StringVar(&dispositionsPath, "dispositions", "", "comment_dispositions array — same input forms as --ops.")
	c.Flags().StringVar(&intent, "intent", "", "Why this version was shipped (audit trail).")
	c.Flags().StringVar(&title, "title", "", "Update the title.")
	c.Flags().StringVar(&summary, "summary", "", "Update the summary.")
	c.Flags().StringSliceVar(&tags, "tag", nil, "Replace tag set (must all exist).")
	c.Flags().StringVar(&actor, "actor", "agent", "Actor type: human | agent.")
	_ = c.MarkFlagRequired("ops")
	return c
}

func pinDocCmd() *cobra.Command {
	var slot int

	c := &cobra.Command{
		Use:   "pin-doc <slug>",
		Short: "Pin a doc to a home-page slot (1-6).",
		Long: `Pin a doc to one of the workspace's 6 numbered home-page slots.
Omit --slot to take the lowest free one. An occupied slot errors with
pin_slot_taken (slots never displace silently); all slots taken errors
with pin_limit_reached. The orientation doc has its own card and can't
be slotted.`,
		Example: `  aveline pin-doc oncall-runbook --slot 2
  aveline pin-doc deploy-guide`,
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
			body := map[string]any{}
			if cmd.Flags().Changed("slot") {
				body["slot"] = slot
			}
			ctx, cancel := withTimeout()
			defer cancel()
			raw, apiErr := client.Post(ctx,
				fmt.Sprintf("/api/workspaces/%s/docs/%s/pin", ws, args[0]), body)
			return handle(raw, apiErr)
		},
	}

	c.Flags().IntVar(&slot, "slot", 0, "Slot number 1-6 (omit for the lowest free slot).")
	return c
}

func unpinDocCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "unpin-doc <slug>",
		Short:        "Free a doc's home-page pin slot.",
		Example:      `  aveline unpin-doc deploy-guide`,
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s/pin", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}

func deleteDocCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "delete-doc <slug>",
		Short:        "Soft-delete a doc.",
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s", ws, args[0]))
			return handle(raw, apiErr)
		},
	}
}

func restoreDocCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "restore-doc <slug>",
		Short:        "Restore a previously soft-deleted doc.",
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s/restore", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

func kudosDocCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "kudos-doc <slug>",
		Short:        "Toggle kudos on a doc (you can't kudos your own).",
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
				fmt.Sprintf("/api/workspaces/%s/docs/%s/kudos", ws, args[0]), nil)
			return handle(raw, apiErr)
		},
	}
}

// ----- Shared helpers --------------------------------------------------

// readBlocks reads a JSON-array argument that may be (a) the path to a
// file, (b) "-" (stdin), or (c) the raw JSON itself.
func readBlocks(arg string) ([]any, error) {
	if arg == "" {
		return nil, fmt.Errorf("--blocks is required")
	}
	v, err := readJSONInput(arg)
	if err != nil {
		return nil, err
	}
	// Accept either a top-level array (blocks list) or an object
	// with a `blocks` key (so users can dump the same JSON they'd
	// PATCH with).
	if arr, ok := v.([]any); ok {
		return arr, nil
	}
	if m, ok := v.(map[string]any); ok {
		if b, ok := m["blocks"].([]any); ok {
			return b, nil
		}
	}
	return nil, fmt.Errorf("expected a JSON array of blocks (or an object with a `blocks` field)")
}

// readJSONInput returns the decoded JSON value from a path, stdin, or
// a raw JSON literal.
func readJSONInput(arg string) (any, error) {
	var data []byte
	var err error
	switch {
	case arg == "-":
		data, err = io.ReadAll(os.Stdin)
	case strings.HasPrefix(strings.TrimSpace(arg), "[") ||
		strings.HasPrefix(strings.TrimSpace(arg), "{"):
		data = []byte(arg)
	case fileExists(arg):
		data, err = os.ReadFile(arg)
	default:
		return nil, fmt.Errorf("not a file, '-', or raw JSON: %q", arg)
	}
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return v, nil
}

func defaultStr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
