package client

import "context"

// ListViews lists all views in a workspace.
func (c *Client) ListViews(ctx context.Context, workspace string) ([]View, error) {
	var out struct {
		Views []View `json:"views"`
	}
	if err := c.Do(ctx, "GET", "/api/workspaces/"+workspace+"/views", nil, nil, &out); err != nil {
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
