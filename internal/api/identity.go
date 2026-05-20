package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// FieldValue safely dereferences an IdentityField pointer, returning the Value
// string or empty string when the field is nil.
func FieldValue(f *IdentityField) string {
	if f == nil {
		return ""
	}
	return f.Value
}

// identityQueryParams builds url.Values from an IdentityListParams.
func identityQueryParams(p IdentityListParams) url.Values {
	q := url.Values{}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if p.SearchType != "" {
		q.Set("searchType", p.SearchType)
	}
	for _, f := range p.Filters {
		q.Add("filters", f)
	}
	if p.Sort != "" {
		q.Set("sort", p.Sort)
	}
	if p.SortDir != "" {
		q.Set("sortDirection", p.SortDir)
	}
	if p.Limit > 0 {
		q.Set("count", strconv.Itoa(p.Limit))
	}
	if p.Cursor != "" {
		q.Set("cursor", p.Cursor)
	}
	return q
}

// ---------------------------------------------------------------------------
// List operations
// ---------------------------------------------------------------------------

// ListUsers fetches a page of identity entities from GET /identity/v1/entities.
func (c *Client) ListUsers(ctx context.Context, params IdentityListParams) (UsersPage, error) {
	q := identityQueryParams(params)
	path := "/identity/v1/entities"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return UsersPage{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return UsersPage{}, err
	}

	var page UsersPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return UsersPage{}, fmt.Errorf("decoding users response: %w", err)
	}
	return page, nil
}

// ListAgents fetches a page of agent entities from GET /identity/v1/agents.
func (c *Client) ListAgents(ctx context.Context, params IdentityListParams) (UsersPage, error) {
	q := identityQueryParams(params)
	path := "/identity/v1/agents"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return UsersPage{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return UsersPage{}, err
	}

	var page UsersPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return UsersPage{}, fmt.Errorf("decoding agents response: %w", err)
	}
	return page, nil
}

// ---------------------------------------------------------------------------
// Single-entity operations
// ---------------------------------------------------------------------------

// GetUser fetches a single identity entity by ID from GET /identity/v1/entities/{id}.
func (c *Client) GetUser(ctx context.Context, id int64) (*Identity, error) {
	path := fmt.Sprintf("/identity/v1/entities/%d", id)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var identity Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("decoding user response: %w", err)
	}
	return &identity, nil
}

// UpdateUser patches scalar fields on an identity via
// PATCH /identity/v1/entities/{id}/revision/{revision}.
func (c *Client) UpdateUser(ctx context.Context, id, revision int64, req UpdateUserRequest) (*Identity, error) {
	path := fmt.Sprintf("/identity/v1/entities/%d/revision/%d", id, revision)

	body := map[string]any{}
	if req.DisplayName != nil {
		body["displayName"] = map[string]string{"value": *req.DisplayName}
	}
	if req.FirstName != nil {
		body["firstName"] = map[string]string{"value": *req.FirstName}
	}
	if req.LastName != nil {
		body["lastName"] = map[string]string{"value": *req.LastName}
	}
	if req.Timezone != nil {
		body["timezone"] = map[string]string{"value": *req.Timezone}
	}
	if req.Tracked != nil {
		body["tracking"] = *req.Tracked
	}

	resp, err := c.doRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var identity Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("decoding update response: %w", err)
	}
	return &identity, nil
}

// DeleteUser deletes an identity entity via DELETE /identity/v1/entities/{id}?revision={revision}.
func (c *Client) DeleteUser(ctx context.Context, id, revision int64) error {
	path := fmt.Sprintf("/identity/v1/entities/%d?revision=%d", id, revision)

	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// ---------------------------------------------------------------------------
// Group membership operations
// ---------------------------------------------------------------------------

// AddUserGroups adds one or more groups to an identity via
// POST /identity/v1/entities/{id}/groups.
func (c *Client) AddUserGroups(ctx context.Context, userID int64, groupIDs []int, revision int64) (*Identity, error) {
	path := fmt.Sprintf("/identity/v1/entities/%d/groups", userID)
	body := map[string]any{
		"groupIds": groupIDs,
		"revision": revision,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var identity Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("decoding add groups response: %w", err)
	}
	return &identity, nil
}

// RemoveUserGroups removes one or more groups from an identity via
// DELETE /identity/v1/entities/{id}/groups.
func (c *Client) RemoveUserGroups(ctx context.Context, userID int64, groupIDs []int, revision int64) (*Identity, error) {
	path := fmt.Sprintf("/identity/v1/entities/%d/groups", userID)
	body := map[string]any{
		"groupIds": groupIDs,
		"revision": revision,
	}

	resp, err := c.doRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var identity Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("decoding remove groups response: %w", err)
	}
	return &identity, nil
}

// ---------------------------------------------------------------------------
// Bulk operations
// ---------------------------------------------------------------------------

// BulkAction applies bulk actions to multiple entities via POST /identity/v1/entities/bulk.
func (c *Client) BulkAction(ctx context.Context, req BulkActionRequest) (*BulkActionResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/identity/v1/entities/bulk", req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var result BulkActionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding bulk action response: %w", err)
	}
	return &result, nil
}

// FetchRevisions concurrently fetches the current revision for each entity ID.
// Returns a map[id]revision and a slice of errors for IDs that could not be fetched.
// At most maxConcurrent GETs run simultaneously.
func (c *Client) FetchRevisions(ctx context.Context, ids []int64, maxConcurrent int) (map[int64]int64, []error) {
	type result struct {
		id  int64
		rev int64
		err error
	}

	sem := make(chan struct{}, maxConcurrent)
	results := make(chan result, len(ids))
	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			identity, err := c.GetUser(ctx, id)
			if err != nil {
				results <- result{id: id, err: err}
				return
			}
			results <- result{id: id, rev: identity.Revision}
		}(id)
	}

	wg.Wait()
	close(results)

	revisions := make(map[int64]int64, len(ids))
	var errs []error
	for r := range results {
		if r.err != nil {
			errs = append(errs, fmt.Errorf("entity %d: %w", r.id, r.err))
		} else {
			revisions[r.id] = r.rev
		}
	}
	return revisions, errs
}

// ParseGroupIDs splits a comma-separated string of integers into []int.
func ParseGroupIDs(s string) ([]int, error) {
	parts := strings.Split(strings.TrimSpace(s), ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid group ID %q: %w", p, err)
		}
		ids = append(ids, n)
	}
	return ids, nil
}
