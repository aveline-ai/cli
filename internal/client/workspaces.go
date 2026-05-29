package client

import "context"

// Me returns the current user plus the workspaces they're a member of.
func (c *Client) Me(ctx context.Context) (*MeResponse, error) {
	var out MeResponse
	if err := c.Do(ctx, "GET", "/api/me", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListWorkspaces returns the workspaces the current user is a member of.
func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	var out struct {
		Workspaces []Workspace `json:"workspaces"`
	}
	if err := c.Do(ctx, "GET", "/api/workspaces", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Workspaces, nil
}

// GetWorkspace returns a single workspace by slug.
func (c *Client) GetWorkspace(ctx context.Context, slug string) (*Workspace, error) {
	var out Workspace
	if err := c.Do(ctx, "GET", "/api/workspaces/"+slug, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
