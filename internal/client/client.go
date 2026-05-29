// Package client speaks the Aveline JSON API.
//
// One Client is created per command invocation, wrapping a *http.Client,
// the resolved API URL, and the bearer token. All requests attach
// Authorization: Bearer <token>. Non-2xx responses are decoded into
// APIError and returned verbatim (the CLI must never reword them).
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the HTTP client for the Aveline API.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// New constructs a Client with a sensible default http.Client.
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// APIError is the decoded form of the server's error envelope.
type APIError struct {
	Status  int            `json:"-"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Field   string         `json:"field,omitempty"`
	Context map[string]any `json:"context,omitempty"`
	// Raw holds the original response body when decoding the envelope
	// fails (e.g. an HTML 502 from a proxy).
	Raw string `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return e.Code
	}
	if e.Raw != "" {
		return fmt.Sprintf("HTTP %d: %s", e.Status, e.Raw)
	}
	return fmt.Sprintf("HTTP %d", e.Status)
}

type errorEnvelope struct {
	Error APIError `json:"error"`
}

// Do executes a request against path (joined with BaseURL), sending body as
// JSON if non-nil, and decoding the JSON response into out if non-nil.
//
// On 204 No Content, out is left untouched.
//
// Non-2xx responses are returned as *APIError.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	if c.Token == "" {
		return errors.New("no API token configured; run `aveline login`")
	}

	u := c.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reqBody = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("calling %s %s: %w", method, u, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if resp.StatusCode == http.StatusNoContent || len(raw) == 0 || out == nil {
			return nil
		}
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
		return nil
	}

	var env errorEnvelope
	if err := json.Unmarshal(raw, &env); err == nil && env.Error.Message != "" {
		env.Error.Status = resp.StatusCode
		return &env.Error
	}
	return &APIError{Status: resp.StatusCode, Raw: strings.TrimSpace(string(raw))}
}
