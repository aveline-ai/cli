package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(handler http.Handler) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	return New(srv.URL, "avl_test"), srv
}

func TestMe(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer avl_test", r.Header.Get("Authorization"))
		assert.Equal(t, "/api/me", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"user":{"id":"u1","username":"arie","email":"a@b","display_name":"Arie"},
			"workspaces":[{"id":"w1","slug":"sp","name":"Pod","inserted_at":"t","updated_at":"t","deleted_at":null}]
		}`)
	}))
	defer srv.Close()

	me, err := c.Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "arie", me.User.Username)
	require.Len(t, me.Workspaces, 1)
	assert.Equal(t, "sp", me.Workspaces[0].Slug)
}

func TestListItemsPropagatesQuery(t *testing.T) {
	var gotQuery string
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"items":[]}`)
	}))
	defer srv.Close()

	_, err := c.ListItems(context.Background(), "sp", ListItemsParams{
		Pinned: true,
		Tags:   []string{"oncall", "ops"},
		View:   "oncall",
	})
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "pinned=true")
	assert.Contains(t, gotQuery, "tag=oncall")
	assert.Contains(t, gotQuery, "tag=ops")
	assert.Contains(t, gotQuery, "view=oncall")
}

func TestUnauthorizedEnvelope(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = io.WriteString(w, `{"error":{"code":"unauthorized","message":"token invalid"}}`)
	}))
	defer srv.Close()

	_, err := c.ListItems(context.Background(), "sp", ListItemsParams{})
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 401, apiErr.Status)
	assert.Equal(t, "unauthorized", apiErr.Code)
	assert.Equal(t, "token invalid", apiErr.Message)
}

func TestCreateItem(t *testing.T) {
	var gotBody map[string]any
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/workspaces/sp/items", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = io.WriteString(w, `{"id":"i1","slug":"hello","title":"Hello","body":"","tags":["t"],"pinned":true}`)
	}))
	defer srv.Close()

	got, err := c.CreateItem(context.Background(), "sp", CreateItemRequest{
		Title: "Hello", Body: "", Tags: []string{"t"}, Pinned: true, Slug: "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, "hello", got.Slug)
	assert.Equal(t, "Hello", gotBody["title"])
	assert.Equal(t, true, gotBody["pinned"])
	assert.Equal(t, "hello", gotBody["slug"])
}

func TestCreateItemValidationFailed(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		_, _ = io.WriteString(w, `{"error":{"code":"slug_taken","message":"slug already in use","field":"slug"}}`)
	}))
	defer srv.Close()

	_, err := c.CreateItem(context.Background(), "sp", CreateItemRequest{Title: "X", Slug: "x"})
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "slug_taken", apiErr.Code)
	assert.Equal(t, "slug", apiErr.Field)
}

func TestDeleteItem204(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	}))
	defer srv.Close()
	require.NoError(t, c.DeleteItem(context.Background(), "sp", "x"))
}

func TestGetViewWithItems(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/workspaces/sp/views/oncall", r.URL.Path)
		_, _ = io.WriteString(w, `{
			"view":{"id":"v","slug":"oncall","name":"Oncall","tag_filter":["oncall"],"description":null,"inserted_at":"t","updated_at":"t","deleted_at":null},
			"items":[{"id":"i1","slug":"a","title":"A","body":"","tags":["oncall"],"pinned":false}]
		}`)
	}))
	defer srv.Close()

	out, err := c.GetView(context.Background(), "sp", "oncall")
	require.NoError(t, err)
	assert.Equal(t, "oncall", out.View.Slug)
	require.Len(t, out.Items, 1)
	assert.Equal(t, "a", out.Items[0].Slug)
}

func TestMissingTokenError(t *testing.T) {
	c := New("http://x", "")
	_, err := c.Me(context.Background())
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "aveline login"))
	// Not an APIError; just a local guard.
	var apiErr *APIError
	assert.False(t, errors.As(err, &apiErr))
}
