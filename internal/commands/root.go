// Package commands wires up the cobra command tree.
package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aveline-ai/cli/internal/client"
	"github.com/aveline-ai/cli/internal/config"
	"github.com/spf13/cobra"
)

// Globals are the persistent flags + lazily-loaded config shared by all
// subcommands. Subcommands access them via cmd.Context().Value(...) — we
// just stash them on the *Globals struct returned by NewRoot.
type Globals struct {
	APIURL    string
	JSONOut   bool
	Workspace string // --workspace override

	// loaded lazily
	cfg       *config.Config
	loadErr   error
	loaded    bool
}

// Config returns the (cached) on-disk config.
func (g *Globals) Config() (config.Config, error) {
	if !g.loaded {
		c, err := config.Load()
		g.cfg = &c
		g.loadErr = err
		g.loaded = true
	}
	return *g.cfg, g.loadErr
}

// Client builds an authenticated API client using effective config.
func (g *Globals) Client() (*client.Client, error) {
	cfg, err := g.Config()
	if err != nil {
		return nil, err
	}
	url := cfg.Resolve(g.APIURL)
	return client.New(url, cfg.Token), nil
}

// WorkspaceSlug returns the workspace to target: flag override > config.
func (g *Globals) WorkspaceSlug() (string, error) {
	if g.Workspace != "" {
		return g.Workspace, nil
	}
	cfg, err := g.Config()
	if err != nil {
		return "", err
	}
	if cfg.Workspace == "" {
		return "", errors.New("no workspace selected; run `aveline workspace use <slug>` or pass --workspace")
	}
	return cfg.Workspace, nil
}

// NewRoot builds the root command and all subcommands. The returned
// *Globals is shared so tests can inspect/override state.
func NewRoot() (*cobra.Command, *Globals) {
	g := &Globals{}

	root := &cobra.Command{
		Use:           "aveline",
		Short:         "Aveline CLI — workspaces, items, and views",
		Long:          "aveline is the command-line client for the Aveline API.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&g.APIURL, "api-url", "", "Aveline API URL (overrides config / AVELINE_API_URL)")
	root.PersistentFlags().BoolVar(&g.JSONOut, "json", false, "emit raw JSON output (machine readable)")
	root.PersistentFlags().StringVar(&g.Workspace, "workspace", "", "workspace slug (overrides config)")

	root.AddCommand(
		newLoginCmd(g),
		newWhoamiCmd(g),
		newWorkspaceCmd(g),
		newListCmd(g),
		newGetCmd(g),
		newSaveCmd(g),
		newEditCmd(g),
		newDeleteCmd(g),
		newRestoreCmd(g),
		newViewCmd(g),
	)
	return root, g
}

// Execute runs the root command and translates errors into exit codes.
func Execute() int {
	root, _ := NewRoot()
	if err := root.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		return 1
	}
	return 0
}

// readBody implements the --body - | FILE convention.
//
//   - flag unset (empty)  → ""
//   - flag == "-"          → read stdin
//   - otherwise            → read file
func readBody(flag string, stdin io.Reader) (string, bool, error) {
	if flag == "" {
		return "", false, nil
	}
	if flag == "-" {
		raw, err := io.ReadAll(stdin)
		if err != nil {
			return "", false, fmt.Errorf("reading stdin: %w", err)
		}
		return string(raw), true, nil
	}
	raw, err := os.ReadFile(flag)
	if err != nil {
		return "", false, fmt.Errorf("reading %s: %w", flag, err)
	}
	return string(raw), true, nil
}
