package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aveline-ai/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupEnv isolates each test's config dir so on-disk state can't leak.
func setupEnv(t *testing.T, cfg config.Config) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, config.Save(cfg))
}

func runCmd(t *testing.T, srvURL string, args ...string) (string, string, error) {
	t.Helper()
	root, _ := NewRoot()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetIn(strings.NewReader(""))
	fullArgs := append([]string{"--api-url", srvURL}, args...)
	root.SetArgs(fullArgs)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

func TestMergeTags(t *testing.T) {
	got := mergeTags([]string{"a", "b", "c"}, []string{"d", "a"}, []string{"b"})
	assert.Equal(t, []string{"a", "c", "d"}, got)

	got = mergeTags(nil, []string{"a", "a"}, nil)
	assert.Equal(t, []string{"a"}, got)

	got = mergeTags([]string{"x"}, nil, []string{"x"})
	assert.Equal(t, []string{}, got)
}

func TestWhoamiJSON(t *testing.T) {
	setupEnv(t, config.Config{Token: "avl_t", Workspace: "sp"})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/me", r.URL.Path)
		assert.Equal(t, "Bearer avl_t", r.Header.Get("Authorization"))
		_, _ = io.WriteString(w, `{"user":{"id":"u","username":"arie","email":"a@b","display_name":null},"workspaces":[]}`)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv.URL, "whoami", "--json")
	require.NoError(t, err)
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	user := got["user"].(map[string]any)
	assert.Equal(t, "arie", user["username"])
	assert.Equal(t, "a@b", user["email"])
}

func TestSaveBuildsRequestBody(t *testing.T) {
	setupEnv(t, config.Config{Token: "avl_t", Workspace: "sp"})

	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = io.WriteString(w, `{"id":"i","slug":"oncall-rotation","title":"Oncall rotation","body":"","tags":["oncall"],"pinned":true}`)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv.URL,
		"save",
		"--title", "Oncall rotation",
		"--tag", "oncall",
		"--pin",
	)
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/api/workspaces/sp/items", gotPath)
	assert.Equal(t, "Oncall rotation", gotBody["title"])
	assert.Equal(t, true, gotBody["pinned"])
	assert.Equal(t, "oncall-rotation", gotBody["slug"])
	tags := gotBody["tags"].([]any)
	require.Len(t, tags, 1)
	assert.Equal(t, "oncall", tags[0])
	assert.Contains(t, stdout, "oncall-rotation")
}

func TestEditAddTagPerformsGetThenPatch(t *testing.T) {
	setupEnv(t, config.Config{Token: "avl_t", Workspace: "sp"})

	var calls []string
	var patchBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch r.Method {
		case "GET":
			_, _ = io.WriteString(w, `{"id":"i","slug":"x","title":"X","body":"","tags":["a","b"],"pinned":false}`)
		case "PATCH":
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &patchBody)
			_, _ = io.WriteString(w, `{"id":"i","slug":"x","title":"X","body":"","tags":["a","c"],"pinned":false}`)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer srv.Close()

	_, _, err := runCmd(t, srv.URL,
		"edit", "x",
		"--add-tag", "c",
		"--remove-tag", "b",
	)
	require.NoError(t, err)
	require.Equal(t, []string{
		"GET /api/workspaces/sp/items/x",
		"PATCH /api/workspaces/sp/items/x",
	}, calls)
	tags := patchBody["tags"].([]any)
	assert.Equal(t, []any{"a", "c"}, tags)
}

func TestErrorEnvelopeRenderedToStderr(t *testing.T) {
	setupEnv(t, config.Config{Token: "avl_t", Workspace: "sp"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"item not found"}}`)
	}))
	defer srv.Close()

	_, _, err := runCmd(t, srv.URL, "get", "missing")
	require.Error(t, err)
	assert.Equal(t, "item not found", err.Error())
}

func TestNoWorkspaceConfigured(t *testing.T) {
	setupEnv(t, config.Config{Token: "avl_t"})
	_, _, err := runCmd(t, "http://unused", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspace selected")
}
