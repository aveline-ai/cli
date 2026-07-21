package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a thin wrapper around net/http. Every method returns either
// the raw JSON payload bytes (caller can json.Decode or just stream) or
// a typed *Error.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// New constructs a Client. baseURL is the API root, e.g. "https://app.aveline.ai".
// token is the bearer token (can be empty for unauthenticated calls like
// /api/heartbeat).
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Do performs an HTTP request against an API path (e.g. "/api/me") and
// returns the raw payload bytes on success or a typed *Error on failure.
//
// `body` can be nil (no payload) or any json.Marshaler-compatible value.
//
// `query` is appended as query parameters when non-empty.
func (c *Client) Do(ctx context.Context, method, path string, body any, query url.Values) ([]byte, *Error) {
	u := c.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &Error{
				Code:    "client_encode_error",
				Message: fmt.Sprintf("encoding request body: %v", err),
			}
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return nil, &Error{
			Code:    "client_request_error",
			Message: err.Error(),
		}
	}

	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, &Error{
			Code:    "network_error",
			Message: err.Error(),
		}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{
			Code:       "client_read_error",
			Message:    err.Error(),
			HTTPStatus: resp.StatusCode,
		}
	}

	return parseEnvelope(resp.StatusCode, raw)
}

// Convenience methods. All return the full envelope bytes (including the
// outer `{ok: true, ...}` wrapper) so callers can print verbatim.

func (c *Client) Get(ctx context.Context, path string, query url.Values) ([]byte, *Error) {
	return c.Do(ctx, http.MethodGet, path, nil, query)
}

func (c *Client) Post(ctx context.Context, path string, body any) ([]byte, *Error) {
	return c.Do(ctx, http.MethodPost, path, body, nil)
}

func (c *Client) Put(ctx context.Context, path string, body any) ([]byte, *Error) {
	return c.Do(ctx, "PUT", path, body, nil)
}

func (c *Client) Patch(ctx context.Context, path string, body any) ([]byte, *Error) {
	return c.Do(ctx, http.MethodPatch, path, body, nil)
}

func (c *Client) Delete(ctx context.Context, path string) ([]byte, *Error) {
	return c.Do(ctx, http.MethodDelete, path, nil, nil)
}
