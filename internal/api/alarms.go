package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListAlarms returns a paginated list of alarms.
// GET /admin/v1/alarms
func (c *Client) ListAlarms(ctx context.Context, page, pageSize int) ([]Alarm, error) {
	q := url.Values{}
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	path := "/admin/v1/alarms?" + q.Encode()

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var alarms []Alarm
	if err := json.NewDecoder(resp.Body).Decode(&alarms); err != nil {
		return nil, fmt.Errorf("decoding alarms response: %w", err)
	}
	return alarms, nil
}

// GetAlarm returns a single alarm by ID.
// GET /admin/v1/alarms/{id}
func (c *Client) GetAlarm(ctx context.Context, id int) (*Alarm, error) {
	path := fmt.Sprintf("/admin/v1/alarms/%d", id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var alarm Alarm
	if err := json.NewDecoder(resp.Body).Decode(&alarm); err != nil {
		return nil, fmt.Errorf("decoding alarm response: %w", err)
	}
	return &alarm, nil
}

// GetAlarmDetails returns the detailed configuration map for an alarm.
// GET /admin/v1/alarmdetails/{id}
func (c *Client) GetAlarmDetails(ctx context.Context, id int) (map[string]any, error) {
	path := fmt.Sprintf("/admin/v1/alarmdetails/%d", id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding alarm details response: %w", err)
	}
	return result, nil
}

// SaveAlarms creates one or more alarms in bulk.
// POST /admin/v1/alarms
func (c *Client) SaveAlarms(ctx context.Context, body map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/alarms", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// SaveAlarm replaces an existing alarm.
// PUT /admin/v1/alarms
func (c *Client) SaveAlarm(ctx context.Context, body map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/admin/v1/alarms", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// DeleteAlarm removes the alarm with the given ID.
// DELETE /admin/v1/alarms/{id}
func (c *Client) DeleteAlarm(ctx context.Context, id int) error {
	path := fmt.Sprintf("/admin/v1/alarms/%d", id)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetAlarmConditions returns the available alarm condition types.
// GET /admin/v1/alarms/conditions
func (c *Client) GetAlarmConditions(ctx context.Context) ([]map[string]any, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/alarms/conditions", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var result []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding alarm conditions response: %w", err)
	}
	return result, nil
}

// GetAlarmFields returns the available alarm field types.
// GET /admin/v1/alarms/fields
func (c *Client) GetAlarmFields(ctx context.Context) ([]map[string]any, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/alarms/fields", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var result []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding alarm fields response: %w", err)
	}
	return result, nil
}
