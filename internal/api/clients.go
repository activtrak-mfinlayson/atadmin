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
// Read operations
// ---------------------------------------------------------------------------

// ListClients returns a paginated slice of clients.
// Page and PageSize are passed as query parameters.
func (c *Client) ListClients(ctx context.Context, page, pageSize int) ([]ATClient, error) {
	q := url.Values{}
	q.Set("Page", strconv.Itoa(page))
	q.Set("PageSize", strconv.Itoa(pageSize))
	path := "/admin/v1/clients?" + q.Encode()

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var clients []ATClient
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, fmt.Errorf("decoding clients response: %w", err)
	}
	return clients, nil
}

// GetClientByID fetches a single client by its numeric ID.
func (c *Client) GetClientByID(ctx context.Context, id int) (*ATClient, error) {
	path := fmt.Sprintf("/admin/v1/clients/%d", id)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var client ATClient
	if err := json.NewDecoder(resp.Body).Decode(&client); err != nil {
		return nil, fmt.Errorf("decoding client response: %w", err)
	}
	return &client, nil
}

// GetClientByUsername fetches a single client by username string.
func (c *Client) GetClientByUsername(ctx context.Context, username string) (*ATClient, error) {
	path := fmt.Sprintf("/admin/v1/clients/%s", username)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var client ATClient
	if err := json.NewDecoder(resp.Body).Decode(&client); err != nil {
		return nil, fmt.Errorf("decoding client response: %w", err)
	}
	return &client, nil
}

// ClientHealth returns the count of currently active clients.
func (c *Client) ClientHealth(ctx context.Context) (int, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/clients/health", nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return 0, err
	}

	var result struct {
		ActiveCount int `json:"activeCount"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decoding health response: %w", err)
	}
	return result.ActiveCount, nil
}

// ---------------------------------------------------------------------------
// Write operations
// ---------------------------------------------------------------------------

// DeleteClients deletes the clients identified by ids.
func (c *Client) DeleteClients(ctx context.Context, ids []int) error {
	body := struct {
		ClientIDs []int `json:"clientIds"`
	}{ClientIDs: ids}

	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/clients", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// RestoreClients restores previously deleted clients.
func (c *Client) RestoreClients(ctx context.Context, ids []int) error {
	body := struct {
		ClientIDs []int `json:"clientIds"`
	}{ClientIDs: ids}

	resp, err := c.doRequest(ctx, http.MethodPut, "/admin/v1/clients/restore", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// UpdateClientAlias sets the display alias for the client with the given id.
func (c *Client) UpdateClientAlias(ctx context.Context, id int, alias string) error {
	path := fmt.Sprintf("/admin/v1/clients/%d", id)
	body := struct {
		Alias string `json:"alias"`
	}{Alias: alias}

	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// MergeUsers merges the source user into the target user.
func (c *Client) MergeUsers(ctx context.Context, sourceID, targetID int) error {
	body := MergeUser{SourceID: sourceID, TargetID: targetID}

	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/clients/mergeusers", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// MergeUsersBulk performs a bulk merge operation from records.
// Each record should contain "sourceId" and "targetId" keys.
func (c *Client) MergeUsersBulk(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/clients/mergeusers/bulk", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// UnmergeUsersBulk reverses previously merged user records in bulk.
func (c *Client) UnmergeUsersBulk(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/clients/unmergeusers/bulk", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// UpdateAlias sets a display alias by the internal user ID (useralias endpoint).
func (c *Client) UpdateAlias(ctx context.Context, id int, alias string) error {
	body := struct {
		ID    int    `json:"id"`
		Alias string `json:"alias"`
	}{ID: id, Alias: alias}

	resp, err := c.doRequest(ctx, http.MethodPut, "/admin/v1/clients/useralias", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// UpdateAliasBulk performs a bulk alias update.
func (c *Client) UpdateAliasBulk(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/clients/useralias/bulk", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// ---------------------------------------------------------------------------
// Do Not Track operations
// ---------------------------------------------------------------------------

// ListDoNotTrack returns all Do Not Track rules for the account.
func (c *Client) ListDoNotTrack(ctx context.Context) ([]DNTEntry, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/clients/donottrack", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var envelope struct {
		Records []DNTEntry `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decoding do-not-track response: %w", err)
	}
	return envelope.Records, nil
}

// AddDoNotTrack adds a single Do Not Track rule for the given domain/username.
func (c *Client) AddDoNotTrack(ctx context.Context, domain, username string) error {
	body := struct {
		LogonDomain string `json:"logonDomain"`
		Username    string `json:"username"`
	}{LogonDomain: domain, Username: username}

	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/clients/donottrack", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// RemoveDoNotTrack deletes Do Not Track rules by their IDs.
func (c *Client) RemoveDoNotTrack(ctx context.Context, ids []int) error {
	body := struct {
		IDs []int `json:"ids"`
	}{IDs: ids}

	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/clients/donottrack", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// UpdateDoNotTrack updates an existing Do Not Track rule.
func (c *Client) UpdateDoNotTrack(ctx context.Context, id int, domain, username string) error {
	body := struct {
		ID          int    `json:"id"`
		LogonDomain string `json:"logonDomain"`
		Username    string `json:"username"`
	}{ID: id, LogonDomain: domain, Username: username}

	resp, err := c.doRequest(ctx, http.MethodPut, "/admin/v1/clients/donottrack", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// AddDoNotTrackBulk adds multiple Do Not Track rules in a single request.
func (c *Client) AddDoNotTrackBulk(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/clients/donottrack/bulk", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// RemoveDoNotTrackBulk deletes multiple Do Not Track rules in a single request.
func (c *Client) RemoveDoNotTrackBulk(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/clients/donottrack/bulk", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// MarkGlobalUser marks the specified client IDs as global (account-wide) Do Not Track entries.
func (c *Client) MarkGlobalUser(ctx context.Context, ids []int) error {
	body := struct {
		IDs []int `json:"ids"`
	}{IDs: ids}

	resp, err := c.doRequest(ctx, http.MethodPatch, "/admin/v1/clients/donottrack/globaluser", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
