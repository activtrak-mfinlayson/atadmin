package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// ListAuditLogs returns audit log entries filtered by the given parameters.
// Any zero-value or empty parameter is omitted from the query string.
func (c *Client) ListAuditLogs(
	ctx context.Context,
	from, to, filters, sortCol string,
	sortDesc bool,
	page, pageSize int,
) ([]AuditLog, error) {
	q := url.Values{}
	if from != "" {
		q.Set("FromDate", from)
	}
	if to != "" {
		q.Set("ToDate", to)
	}
	if filters != "" {
		q.Set("Filters", filters)
	}
	if sortCol != "" {
		q.Set("SortColumn", sortCol)
	}
	if sortDesc {
		q.Set("SortDescending", "true")
	}
	if page != 0 {
		q.Set("Page", strconv.Itoa(page))
	}
	if pageSize != 0 {
		q.Set("PageSize", strconv.Itoa(pageSize))
	}

	path := "/admin/v1/auditlog"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var logs []AuditLog
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, fmt.Errorf("decoding audit log response: %w", err)
	}
	return logs, nil
}

// GetAttachment retrieves the raw bytes of an audit log attachment by its ID.
func (c *Client) GetAttachment(ctx context.Context, attachmentID string) ([]byte, error) {
	path := fmt.Sprintf("/admin/v1/attachment/%s", attachmentID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading attachment body: %w", err)
	}
	return data, nil
}
