// Package output handles success / failure rendering for every CLI verb.
//
// The CLI is agent-first: JSON is the default. The agent reads
// stdout directly; humans can pass --human at the cobra layer.
//
// On failure, the API's error envelope is rendered verbatim to STDERR
// and the process exits non-zero. The format is the same as the API
// returns:
//
//	{"ok": false, "error": {"code": "...", "message": "..."}}
//
// so an agent can branch on the exit code, then parse stderr for the
// machine-readable code.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aveline-ai/cli/internal/api"
)

// PrintSuccess writes the raw envelope bytes returned by the API to
// stdout, exactly as the server sent them.
func PrintSuccess(raw []byte) error {
	if len(raw) == 0 {
		return nil
	}
	// Always emit a trailing newline so it composes with `| jq` etc.
	if _, err := os.Stdout.Write(raw); err != nil {
		return err
	}
	if raw[len(raw)-1] != '\n' {
		fmt.Fprintln(os.Stdout)
	}
	return nil
}

// PrintError emits the failure envelope to STDERR and returns a usable
// exit code. The agent gets the same JSON shape it would have from
// curl. We never log retry hints — those go in the API's `error.message`.
func PrintError(err *api.Error) int {
	if err == nil {
		return 0
	}
	emit := map[string]any{
		"ok": false,
		"error": map[string]any{
			"code":    err.Code,
			"message": err.Message,
		},
	}
	if err.Details != nil && len(err.Details) > 0 {
		emit["error"].(map[string]any)["details"] = err.Details
	}
	b, _ := json.Marshal(emit)
	fmt.Fprintln(os.Stderr, string(b))

	// Status-based exit code makes the agent able to branch in shell:
	// 0 = success, 2 = validation/business error, 3 = auth, 4 = not
	// found, 1 = anything else (network, encoding).
	switch err.HTTPStatus {
	case 401, 403:
		return 3
	case 404:
		return 4
	case 422:
		return 2
	default:
		return 1
	}
}

// HumanWriter writes pretty-printed JSON for the --human flag. We don't
// have a full TUI yet; pretty-print is the human escape hatch.
func PrintHuman(w io.Writer, raw []byte) error {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		_, err = w.Write(raw)
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
