package client

import (
	"context"
	"net/url"
)

// ListViewsParams encodes optional query params for ListViews.
type ListViewsParams struct {
	Scope string // "personal", "team", or "" for both
}

// ListViews lists all views visible to the caller in a workspace.
func (c *Client) ListViews(ctx context.Context, workspace string, p ListViewsParams) ([]View, error) {
	q := url.Values{}
	if p.Scope != "" {
		q.Set("scope", p.Scope)
	}
	var out struct {
		Views []View `json:"views"`
	}
	if err := c.Do(ctx, "GET", "/api/workspaces/"+workspace+"/views", q, nil, &out); err != nil {
		return nil, err
	}
	return out.Views, nil
}

// GetView fetches a view and its matching items.
func (c *Client) GetView(ctx context.Context, workspace, slug string) (*ViewWithItems, error) {
	var out ViewWithItems
	if err := c.Do(ctx, "GET", "/api/workspaces/"+workspace+"/views/"+slug, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateViewRequest is the POST body for creating a view.
type CreateViewRequest struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	TagFilter   []string `json:"tag_filter"`
	Description *string  `json:"description,omitempty"`
	Scope       string   `json:"scope,omitempty"`
}

// CreateView creates a new saved view.
func (c *Client) CreateView(ctx context.Context, workspace string, req CreateViewRequest) (*View, error) {
	if req.TagFilter == nil {
		req.TagFilter = []string{}
	}
	var out View
	if err := c.Do(ctx, "POST", "/api/workspaces/"+workspace+"/views", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateViewRequest is the PATCH body; nil fields are omitted.
type UpdateViewRequest struct {
	Name        *string   `json:"name,omitempty"`
	TagFilter   *[]string `json:"tag_filter,omitempty"`
	Description *string   `json:"description,omitempty"`
	Scope       *string   `json:"scope,omitempty"`
}

// UpdateView applies a partial update.
func (c *Client) UpdateView(ctx context.Context, workspace, slug string, req UpdateViewRequest) (*View, error) {
	var out View
	if err := c.Do(ctx, "PATCH", "/api/workspaces/"+workspace+"/views/"+slug, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteView soft-deletes a view.
func (c *Client) DeleteView(ctx context.Context, workspace, slug string) error {
	return c.Do(ctx, "DELETE", "/api/workspaces/"+workspace+"/views/"+slug, nil, nil, nil)
}
