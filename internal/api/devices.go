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
// Device read operations
// ---------------------------------------------------------------------------

// ListDevices returns a paginated slice of devices.
func (c *Client) ListDevices(ctx context.Context, page, pageSize int) ([]Device, error) {
	q := url.Values{}
	q.Set("Page", strconv.Itoa(page))
	q.Set("PageSize", strconv.Itoa(pageSize))
	path := "/admin/v1/devices?" + q.Encode()

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var devices []Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("decoding devices response: %w", err)
	}
	return devices, nil
}

// GetDevice fetches a single device by its numeric ID.
func (c *Client) GetDevice(ctx context.Context, id int) (*Device, error) {
	path := fmt.Sprintf("/admin/v1/devices/%d", id)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var device Device
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return nil, fmt.Errorf("decoding device response: %w", err)
	}
	return &device, nil
}

// ---------------------------------------------------------------------------
// Device write operations
// ---------------------------------------------------------------------------

// DeleteDevices deletes the devices identified by ids.
func (c *Client) DeleteDevices(ctx context.Context, ids []int) error {
	body := struct {
		DeviceIDs []int `json:"deviceIds"`
	}{DeviceIDs: ids}

	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/devices", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// RestoreDevices restores previously deleted devices.
func (c *Client) RestoreDevices(ctx context.Context, ids []int) error {
	body := struct {
		DeviceIDs []int `json:"deviceIds"`
	}{DeviceIDs: ids}

	resp, err := c.doRequest(ctx, http.MethodPut, "/admin/v1/devices/restore", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// UninstallDevice sends an uninstall request for the specified device IDs.
func (c *Client) UninstallDevice(ctx context.Context, ids []int) error {
	body := struct {
		DeviceIDs []int `json:"deviceIds"`
	}{DeviceIDs: ids}

	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/devices/uninstall", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
