package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListSignals returns all signals configured for the account.
// GET /admin/v1/signals
func (c *Client) ListSignals(ctx context.Context) ([]Signal, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/signals", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var signals []Signal
	if err := json.NewDecoder(resp.Body).Decode(&signals); err != nil {
		return nil, fmt.Errorf("decoding signals response: %w", err)
	}
	return signals, nil
}

// GetLegacyNotifications returns signals from the legacy notifications endpoint.
// GET /admin/legacy/notifications
func (c *Client) GetLegacyNotifications(ctx context.Context) ([]Signal, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/legacy/notifications", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var signals []Signal
	if err := json.NewDecoder(resp.Body).Decode(&signals); err != nil {
		return nil, fmt.Errorf("decoding legacy notifications response: %w", err)
	}
	return signals, nil
}

// CreateSignal creates a new signal and returns its assigned ID.
// POST /admin/v1/signal
func (c *Client) CreateSignal(ctx context.Context, body map[string]any) (int, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/signal", body)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return 0, err
	}
	var result struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decoding create signal response: %w", err)
	}
	return result.ID, nil
}

// UpdateSignal replaces an existing signal.
// PUT /admin/v1/signal
func (c *Client) UpdateSignal(ctx context.Context, body map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/admin/v1/signal", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// DeleteSignal removes the signal with the given ID.
// DELETE /admin/v1/signals/{id}
func (c *Client) DeleteSignal(ctx context.Context, id int) error {
	path := fmt.Sprintf("/admin/v1/signals/%d", id)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
