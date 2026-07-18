// Package cmd hosts all CLI commands. Verbs are FLAT (`edit-comment`
// rather than `comment edit`) because the primary caller is an LLM
// agent and a flat namespace means one `--help` call reveals the
// entire surface area. See README.md and CLAUDE.md for the philosophy.
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aveline-ai/cli/internal/api"
	"github.com/aveline-ai/cli/internal/config"
	"github.com/aveline-ai/cli/internal/output"
	"github.com/spf13/cobra"
)

// Globals set from cobra flags / config — re-resolved per command.
var (
	flagAPIURL    string
	flagWorkspace string
	flagHuman     bool
)

// Root returns the root cobra.Command with every verb registered.
func Root(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "aveline",
		Short: "Aveline CLI",
		Long: `Aveline CLI — a flat-verb command-line client for the Aveline wiki.

Aveline is a structured wiki designed to be written and read by AI
agents. This CLI is the primary surface agents use to interact with
it.

Verbs are FLAT (not nested). One ` + "`aveline --help`" + ` call surfaces
every operation. Each verb has a ` + "`--help`" + ` with examples that match
how an agent would call it.

Output is JSON by default — pass --human for pretty-print. Errors are
written to stderr as the API's machine-readable error envelope; the
process exit code groups them (2 = validation, 3 = auth, 4 = not
found, 1 = network / other).`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "Override the API URL (otherwise from $AVELINE_API_URL or config).")
	root.PersistentFlags().StringVarP(&flagWorkspace, "workspace", "w", "", "Workspace slug to target. Overrides the config's default.")
	root.PersistentFlags().BoolVar(&flagHuman, "human", false, "Pretty-print JSON for humans. Default is the raw API envelope (for agents).")

	// Auth / session
	root.AddCommand(loginCmd())
	root.AddCommand(logoutCmd())
	root.AddCommand(whoamiCmd())

	// Workspaces
	root.AddCommand(listWorkspacesCmd())
	root.AddCommand(getWorkspaceCmd())
	root.AddCommand(createWorkspaceCmd())
	root.AddCommand(useWorkspaceCmd())

	// Docs
	root.AddCommand(listDocsCmd())
	root.AddCommand(getDocCmd())
	root.AddCommand(getOrientationCmd())
	root.AddCommand(contractCmd())
	root.AddCommand(createDocCmd())
	root.AddCommand(editDocCmd())
	root.AddCommand(pinDocCmd())
	root.AddCommand(unpinDocCmd())
	root.AddCommand(deleteDocCmd())
	root.AddCommand(restoreDocCmd())
	root.AddCommand(kudosDocCmd())
	root.AddCommand(runBlockCmd())

	// Versions
	root.AddCommand(listVersionsCmd())
	root.AddCommand(getVersionCmd())

	// Comments
	root.AddCommand(listCommentsCmd())
	root.AddCommand(createCommentCmd())
	root.AddCommand(replyCommentCmd())
	root.AddCommand(editCommentCmd())
	root.AddCommand(deleteCommentCmd())
	root.AddCommand(undeleteCommentCmd())
	root.AddCommand(resolveCommentCmd())
	root.AddCommand(unresolveCommentCmd())

	// Tags
	root.AddCommand(listTagsCmd())
	root.AddCommand(getTagCmd())
	root.AddCommand(createTagCmd())
	root.AddCommand(editTagCmd())
	root.AddCommand(deleteTagCmd())
	root.AddCommand(restoreTagCmd())

	// Views
	root.AddCommand(createViewCmd())
	root.AddCommand(editViewCmd())
	root.AddCommand(listViewsCmd())
	root.AddCommand(deleteViewCmd())
	root.AddCommand(restoreViewCmd())
	root.AddCommand(pinViewCmd())
	root.AddCommand(unpinViewCmd())

	// Data sources
	root.AddCommand(createDataSourceCmd())
	root.AddCommand(editDataSourceCmd())
	root.AddCommand(queryDataSourceCmd())
	root.AddCommand(listDataSourcesCmd())
	root.AddCommand(deleteDataSourceCmd())

	// Query catalog
	root.AddCommand(createQueryCmd())
	root.AddCommand(editQueryCmd())
	root.AddCommand(listQueriesCmd())
	root.AddCommand(getQueryCmd())
	root.AddCommand(deleteQueryCmd())
	root.AddCommand(restoreQueryCmd())

	// Team
	root.AddCommand(listMembersCmd())
	root.AddCommand(addMemberCmd())
	root.AddCommand(removeMemberCmd())
	root.AddCommand(getInviteCmd())
	root.AddCommand(revokeInviteCmd())

	// Activity
	root.AddCommand(listEventsCmd())

	// Misc
	root.AddCommand(heartbeatCmd())

	return root
}

// Execute is the main entry point.
func Execute(version string) {
	root := Root(version)
	if err := root.Execute(); err != nil {
		// Errors are already rendered by handle() via PrintError. The
		// returned error here is usually a cobra usage error. Still
		// exit non-zero.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ----- Shared helpers used by every verb's RunE ---------------------

// loadCfg reads ~/.config/aveline/config.toml, returning an empty
// config if missing.
func loadCfg() config.Config {
	cfg, _ := config.Load()
	return cfg
}

// resolveAPI builds an api.Client from config + flags. Most verbs
// need this. unauthenticated=true skips the token check.
func resolveAPI(unauthenticated bool) (*api.Client, error) {
	cfg := loadCfg()
	url := cfg.Resolve(flagAPIURL)
	token := cfg.Token
	if !unauthenticated && token == "" {
		return nil, fmt.Errorf("not logged in — run `aveline login` to paste an API key")
	}
	return api.New(url, token), nil
}

// resolveWorkspaceSlug returns the workspace slug to use, in order:
// --workspace flag > config > error.
func resolveWorkspaceSlug() (string, error) {
	if flagWorkspace != "" {
		return flagWorkspace, nil
	}
	cfg := loadCfg()
	if cfg.Workspace != "" {
		return cfg.Workspace, nil
	}
	return "", fmt.Errorf("no workspace configured — run `aveline use-workspace <slug>` or pass --workspace=<slug>")
}

// handle runs a request and prints the result. Use this in every RunE
// so success and failure go to the right streams with the right exit
// code. cobra returns nil from RunE and the process exits with the
// code we set via os.Exit on failure.
func handle(raw []byte, err *api.Error) error {
	if err != nil {
		code := output.PrintError(err)
		os.Exit(code)
	}
	if flagHuman {
		return output.PrintHuman(os.Stdout, raw)
	}
	return output.PrintSuccess(raw)
}

// withTimeout gives every call a sane upper bound so a hung server
// can't lock the agent.
func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 60*time.Second)
}

// fileExists is a tiny helper used by the doc commands when reading
// `--blocks` from a path.
func fileExists(p string) bool {
	if p == "" {
		return false
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return false
	}
	if info, err := os.Stat(abs); err == nil && !info.IsDir() {
		return true
	}
	return false
}
