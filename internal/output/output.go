// Package output formats command results for either human consumption or
// raw JSON (--json), which Claude uses for machine parsing.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aveline-ai/cli/internal/client"
)

// JSON marshals v as indented JSON to w.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Me prints the user + workspaces human-style.
func Me(w io.Writer, me *client.MeResponse) {
	name := me.User.Username
	if me.User.DisplayName != nil && *me.User.DisplayName != "" {
		name = *me.User.DisplayName + " (" + me.User.Username + ")"
	}
	fmt.Fprintf(w, "%s <%s>\n", name, me.User.Email)
	if len(me.Workspaces) == 0 {
		fmt.Fprintln(w, "  (no workspaces)")
		return
	}
	fmt.Fprintln(w, "workspaces:")
	for _, ws := range me.Workspaces {
		fmt.Fprintf(w, "  %s  %s\n", ws.Slug, ws.Name)
	}
}

// Workspaces prints a workspace list.
func Workspaces(w io.Writer, list []client.Workspace) {
	if len(list) == 0 {
		fmt.Fprintln(w, "(no workspaces)")
		return
	}
	for _, ws := range list {
		fmt.Fprintf(w, "%-24s  %s\n", ws.Slug, ws.Name)
	}
}

// Items prints an item list.
func Items(w io.Writer, items []client.Item) {
	if len(items) == 0 {
		fmt.Fprintln(w, "(no items)")
		return
	}
	for _, it := range items {
		pin := " "
		if it.Pinned {
			pin = "*"
		}
		tags := ""
		if len(it.Tags) > 0 {
			tags = "  [" + strings.Join(it.Tags, ", ") + "]"
		}
		fmt.Fprintf(w, "%s %-32s  %s%s\n", pin, it.Slug, it.Title, tags)
	}
}

// Item prints item metadata followed by the body (matching `aveline get`).
func Item(w io.Writer, it *client.Item) {
	fmt.Fprintf(w, "# %s\n", it.Title)
	fmt.Fprintf(w, "slug: %s\n", it.Slug)
	if len(it.Tags) > 0 {
		fmt.Fprintf(w, "tags: %s\n", strings.Join(it.Tags, ", "))
	}
	if it.Pinned {
		fmt.Fprintln(w, "pinned: true")
	}
	if it.Summary != nil && *it.Summary != "" {
		fmt.Fprintf(w, "summary: %s\n", *it.Summary)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, it.Body)
}

// Views prints a view list. Scope shown as a small tag.
func Views(w io.Writer, list []client.View) {
	if len(list) == 0 {
		fmt.Fprintln(w, "(no views)")
		return
	}
	for _, v := range list {
		filter := ""
		if len(v.TagFilter) > 0 {
			filter = "  [" + strings.Join(v.TagFilter, ", ") + "]"
		}
		scope := v.Scope
		if scope == "" {
			scope = "?"
		}
		fmt.Fprintf(w, "%-24s  %-8s  %s%s\n", v.Slug, "("+scope+")", v.Name, filter)
	}
}

// ViewWithItems prints a view plus its items.
func ViewWithItems(w io.Writer, v *client.ViewWithItems) {
	scope := v.View.Scope
	if scope == "" {
		scope = "?"
	}
	fmt.Fprintf(w, "# %s (%s) — %s\n", v.View.Name, v.View.Slug, scope)
	if len(v.View.TagFilter) > 0 {
		fmt.Fprintf(w, "tag_filter: %s\n", strings.Join(v.View.TagFilter, ", "))
	}
	if v.View.Description != nil && *v.View.Description != "" {
		fmt.Fprintf(w, "description: %s\n", *v.View.Description)
	}
	fmt.Fprintln(w)
	Items(w, v.Items)
}

// View prints a single view's metadata.
func View(w io.Writer, v *client.View) {
	fmt.Fprintf(w, "%s\t%s\n", v.Slug, v.Name)
	if len(v.TagFilter) > 0 {
		fmt.Fprintf(w, "tag_filter: %s\n", strings.Join(v.TagFilter, ", "))
	}
	if v.Description != nil && *v.Description != "" {
		fmt.Fprintf(w, "description: %s\n", *v.Description)
	}
}
