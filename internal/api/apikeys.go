package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListAPIKeys returns all API keys for the account.
func (c *Client) ListAPIKeys(ctx context.Context) ([]ApiKey, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/key", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var keys []ApiKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("decoding api keys response: %w", err)
	}
	return keys, nil
}

// CreateAPIKey creates a new API key with the given name and returns the created key.
func (c *Client) CreateAPIKey(ctx context.Context, name string) (*ApiKey, error) {
	body := struct {
		Name string `json:"name"`
	}{Name: name}

	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/key", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var key ApiKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, fmt.Errorf("decoding create api key response: %w", err)
	}
	return &key, nil
}

// UpdateAPIKey updates the name of the API key with the given id.
func (c *Client) UpdateAPIKey(ctx context.Context, id int, name string) error {
	body := struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{ID: id, Name: name}

	resp, err := c.doRequest(ctx, http.MethodPut, "/admin/v1/key", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// DeleteAPIKey deletes the API key with the given id.
func (c *Client) DeleteAPIKey(ctx context.Context, id int) error {
	path := fmt.Sprintf("/admin/v1/key/%d", id)

	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// BackfillAllAPIKeys triggers backfill of instance IDs for all API keys.
func (c *Client) BackfillAllAPIKeys(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/util/backfill_apikey_instanceid", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// BackfillAPIKey triggers backfill of the instance ID for a single API key.
func (c *Client) BackfillAPIKey(ctx context.Context, id int) error {
	path := fmt.Sprintf("/admin/v1/util/backfill_apikey_instanceid/%d", id)

	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
