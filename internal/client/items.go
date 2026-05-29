package client

import (
	"context"
	"net/url"
)

// ListItemsParams encodes optional query params for ListItems.
type ListItemsParams struct {
	Pinned bool
	Tags   []string
	View   string
}

// ListItems lists items in a workspace.
func (c *Client) ListItems(ctx context.Context, workspace string, p ListItemsParams) ([]Item, error) {
	q := url.Values{}
	if p.Pinned {
		q.Set("pinned", "true")
	}
	for _, t := range p.Tags {
		q.Add("tag", t)
	}
	if p.View != "" {
		q.Set("view", p.View)
	}
	var out struct {
		Items []Item `json:"items"`
	}
	if err := c.Do(ctx, "GET", "/api/workspaces/"+workspace+"/items", q, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// GetItem fetches a single item.
func (c *Client) GetItem(ctx context.Context, workspace, slug string) (*Item, error) {
	var out Item
	if err := c.Do(ctx, "GET", "/api/workspaces/"+workspace+"/items/"+slug, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateItemRequest is the POST body for creating an item.
type CreateItemRequest struct {
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	Summary *string  `json:"summary,omitempty"`
	Tags    []string `json:"tags"`
	Pinned  bool     `json:"pinned"`
	Slug    string   `json:"slug,omitempty"`
}

// CreateItem creates a new item.
func (c *Client) CreateItem(ctx context.Context, workspace string, req CreateItemRequest) (*Item, error) {
	if req.Tags == nil {
		req.Tags = []string{}
	}
	var out Item
	if err := c.Do(ctx, "POST", "/api/workspaces/"+workspace+"/items", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateItemRequest is the PATCH body; nil fields are omitted.
type UpdateItemRequest struct {
	Title   *string   `json:"title,omitempty"`
	Body    *string   `json:"body,omitempty"`
	Summary *string   `json:"summary,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
	Pinned  *bool     `json:"pinned,omitempty"`
}

// UpdateItem applies a partial update.
func (c *Client) UpdateItem(ctx context.Context, workspace, slug string, req UpdateItemRequest) (*Item, error) {
	var out Item
	if err := c.Do(ctx, "PATCH", "/api/workspaces/"+workspace+"/items/"+slug, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteItem soft-deletes an item.
func (c *Client) DeleteItem(ctx context.Context, workspace, slug string) error {
	return c.Do(ctx, "DELETE", "/api/workspaces/"+workspace+"/items/"+slug, nil, nil, nil)
}

// RestoreItem undoes a soft-delete.
func (c *Client) RestoreItem(ctx context.Context, workspace, slug string) (*Item, error) {
	var out Item
	if err := c.Do(ctx, "POST", "/api/workspaces/"+workspace+"/items/"+slug+"/restore", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
