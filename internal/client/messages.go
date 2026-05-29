package client

import "context"

// ListMessages returns all messages on an item, oldest first.
func (c *Client) ListMessages(ctx context.Context, workspace, itemSlug string) ([]Message, error) {
	var out struct {
		Messages []Message `json:"messages"`
	}
	if err := c.Do(ctx, "GET", "/api/workspaces/"+workspace+"/items/"+itemSlug+"/messages", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Messages, nil
}

// CreateMessageRequest is the POST body for posting a reply.
type CreateMessageRequest struct {
	Body       string `json:"body"`
	CreatedVia string `json:"created_via,omitempty"`
}

// CreateMessage posts a reply to an item.
func (c *Client) CreateMessage(ctx context.Context, workspace, itemSlug string, req CreateMessageRequest) (*Message, error) {
	if req.CreatedVia == "" {
		req.CreatedVia = "cli"
	}
	var out Message
	if err := c.Do(ctx, "POST", "/api/workspaces/"+workspace+"/items/"+itemSlug+"/messages", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMessageRequest is the PATCH body.
type UpdateMessageRequest struct {
	Body string `json:"body"`
}

// UpdateMessage edits an existing message.
func (c *Client) UpdateMessage(ctx context.Context, workspace, itemSlug, id string, req UpdateMessageRequest) (*Message, error) {
	var out Message
	if err := c.Do(ctx, "PATCH", "/api/workspaces/"+workspace+"/items/"+itemSlug+"/messages/"+id, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMessage soft-deletes a message.
func (c *Client) DeleteMessage(ctx context.Context, workspace, itemSlug, id string) error {
	return c.Do(ctx, "DELETE", "/api/workspaces/"+workspace+"/items/"+itemSlug+"/messages/"+id, nil, nil, nil)
}
