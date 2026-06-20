// Package api speaks Aveline's HTTP API. Every response is shaped as
//
//	{"ok": true, ...payload}                          // success
//	{"ok": false, "error": {"code", "message", ...}}  // failure
//
// Client wraps net/http and unwraps that envelope, returning either
// the raw payload bytes (for the agent to read) or a typed *Error.
//
// The CLI is *agent-first*: callers default to printing the raw
// JSON envelope so Claude can parse it directly. Humans can pass
// --human at the cobra layer.
package api

import (
	"encoding/json"
	"fmt"
)

// Error is the parsed `error` object from a failure envelope. Callers
// branch on Code (machine-readable) and surface Message (human line)
// to the user.
type Error struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]any         `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Raw        map[string]any         `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// envelope is the partial decoding used to detect ok/error before we
// hand the full payload back to the caller.
type envelope struct {
	OK    bool             `json:"ok"`
	Error *json.RawMessage `json:"error,omitempty"`
}

// parseEnvelope decodes raw JSON bytes into either a parsed Error
// (on failure) or returns the raw bytes (on success) for the caller
// to use directly.
func parseEnvelope(status int, body []byte) ([]byte, *Error) {
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, &Error{
			Code:       "client_decode_error",
			Message:    fmt.Sprintf("server returned non-JSON %d: %s", status, truncate(body, 200)),
			HTTPStatus: status,
		}
	}

	if env.OK {
		return body, nil
	}

	// failure path
	if env.Error == nil {
		return nil, &Error{
			Code:       "internal_error",
			Message:    "server returned ok=false without an error object",
			HTTPStatus: status,
		}
	}

	var errObj struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Details map[string]any `json:"details,omitempty"`
	}
	if err := json.Unmarshal(*env.Error, &errObj); err != nil {
		return nil, &Error{
			Code:       "client_decode_error",
			Message:    err.Error(),
			HTTPStatus: status,
		}
	}

	// Preserve the raw envelope so the agent gets the same JSON it
	// would have gotten from curl when we re-emit on stderr.
	var raw map[string]any
	_ = json.Unmarshal(body, &raw)

	return nil, &Error{
		Code:       errObj.Code,
		Message:    errObj.Message,
		Details:    errObj.Details,
		HTTPStatus: status,
		Raw:        raw,
	}
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
