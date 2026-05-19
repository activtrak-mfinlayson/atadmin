package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Consumer read operations
// ---------------------------------------------------------------------------

// ListConsumers returns a paginated slice of consumers.
func (c *Client) ListConsumers(ctx context.Context, page, pageSize int) ([]Consumer, error) {
	q := url.Values{}
	q.Set("Page", strconv.Itoa(page))
	q.Set("PageSize", strconv.Itoa(pageSize))
	path := "/admin/v1/consumers?" + q.Encode()

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var consumers []Consumer
	if err := json.NewDecoder(resp.Body).Decode(&consumers); err != nil {
		return nil, fmt.Errorf("decoding consumers response: %w", err)
	}
	return consumers, nil
}

// GetConsumer fetches a single consumer by its numeric ID.
func (c *Client) GetConsumer(ctx context.Context, id int) (*Consumer, error) {
	path := fmt.Sprintf("/admin/v1/consumers/%d", id)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var consumer Consumer
	if err := json.NewDecoder(resp.Body).Decode(&consumer); err != nil {
		return nil, fmt.Errorf("decoding consumer response: %w", err)
	}
	return &consumer, nil
}

// ---------------------------------------------------------------------------
// Consumer write operations
// ---------------------------------------------------------------------------

// CreateConsumers creates consumers from structured records in bulk.
func (c *Client) CreateConsumers(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/consumers", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// PatchConsumers updates consumers from structured records in bulk.
func (c *Client) PatchConsumers(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPatch, "/admin/v1/consumers", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// DeleteConsumers deletes consumers identified by ids.
func (c *Client) DeleteConsumers(ctx context.Context, ids []int) error {
	body := struct {
		IDs []int `json:"ids"`
	}{IDs: ids}

	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/consumers", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// DeleteConsumersBulk deletes consumers from structured records in bulk.
func (c *Client) DeleteConsumersBulk(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/consumers/bulk", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// SetConsumerRole sets the role for the consumer with the given id.
func (c *Client) SetConsumerRole(ctx context.Context, id int, role string) error {
	path := fmt.Sprintf("/admin/v1/consumers/%d/role", id)
	body := struct {
		Role string `json:"role"`
	}{Role: role}

	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// SetConsumerPassword sets the password for the consumer with the given id.
func (c *Client) SetConsumerPassword(ctx context.Context, id int, password string) error {
	path := fmt.Sprintf("/admin/v1/consumers/%d/password", id)
	body := struct {
		Password string `json:"password"`
	}{Password: password}

	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// SetConsumerSSO enables or disables SSO for the consumer with the given id.
func (c *Client) SetConsumerSSO(ctx context.Context, id int, useSSO bool) error {
	body := struct {
		ConsumerID int  `json:"consumerId"`
		UseSSO     bool `json:"useSSO"`
	}{ConsumerID: id, UseSSO: useSSO}

	resp, err := c.doRequest(ctx, http.MethodPut, "/admin/v1/consumers/usesso", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// AddConsumerViewableGroups grants the consumer visibility into additional groups.
func (c *Client) AddConsumerViewableGroups(ctx context.Context, id int, groupIDs []int) error {
	path := fmt.Sprintf("/admin/v1/consumers/%d/viewablegroups", id)
	body := struct {
		GroupIDs []int `json:"groupIds"`
	}{GroupIDs: groupIDs}

	resp, err := c.doRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// CreateChromeUsersBulk creates Chrome-managed users in bulk.
func (c *Client) CreateChromeUsersBulk(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/consumers/chromeusers/bulk", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
