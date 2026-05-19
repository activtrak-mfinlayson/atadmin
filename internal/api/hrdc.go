package api

import (
	"context"
	"net/http"
)

// HRDCPing checks the health of the HRDC integration endpoint.
func (c *Client) HRDCPing(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodGet, "/hrdc/ping", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// HRDCBulkImport submits a bulk import of HRDC records.
func (c *Client) HRDCBulkImport(ctx context.Context, records []map[string]any) error {
	body := struct {
		Records []map[string]any `json:"records"`
	}{Records: records}

	resp, err := c.doRequest(ctx, http.MethodPost, "/hrdc/v1/bulk", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
