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
// Group read operations
// ---------------------------------------------------------------------------

// groupListItem is the actual wire shape returned by /admin/v1/groups/list.
// The API uses "groupid"/"groupname" instead of "id"/"name".
type groupListItem struct {
	ID   int    `json:"groupid"`
	Name string `json:"groupname"`
}

func (g groupListItem) toGroup() Group {
	return Group{ID: g.ID, Name: g.Name}
}

func decodeGroupList(resp *http.Response) ([]Group, error) {
	var raw []groupListItem
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]Group, len(raw))
	for i, r := range raw {
		out[i] = r.toGroup()
	}
	return out, nil
}

// ListGroups returns a paginated slice of groups.
func (c *Client) ListGroups(ctx context.Context, page, pageSize int) ([]Group, error) {
	q := url.Values{}
	q.Set("Page", strconv.Itoa(page))
	q.Set("PageSize", strconv.Itoa(pageSize))
	path := "/admin/v1/groups/list?" + q.Encode()

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	groups, err := decodeGroupList(resp)
	if err != nil {
		return nil, fmt.Errorf("decoding groups response: %w", err)
	}
	return groups, nil
}

// GetGroupSummary returns a summary slice of all groups.
func (c *Client) GetGroupSummary(ctx context.Context) ([]Group, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/groups/summary", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var envelope struct {
		Groups []Group `json:"groups"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decoding groups summary response: %w", err)
	}
	return envelope.Groups, nil
}

// GetGroup fetches a single group by its numeric ID.
func (c *Client) GetGroup(ctx context.Context, id int) (*Group, error) {
	path := fmt.Sprintf("/admin/v1/groups/list/%d", id)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	groups, err := decodeGroupList(resp)
	if err != nil {
		return nil, fmt.Errorf("decoding group response: %w", err)
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("not found: the requested resource does not exist")
	}
	return &groups[0], nil
}

// SearchGroups returns groups whose name starts with the given prefix.
func (c *Client) SearchGroups(ctx context.Context, prefix string) ([]Group, error) {
	path := fmt.Sprintf("/admin/v1/groups/list/%s", prefix)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	groups, err := decodeGroupList(resp)
	if err != nil {
		return nil, fmt.Errorf("decoding groups search response: %w", err)
	}
	return groups, nil
}

// ---------------------------------------------------------------------------
// Group write operations
// ---------------------------------------------------------------------------

// CreateGroup creates a new group with the given name and returns its ID.
func (c *Client) CreateGroup(ctx context.Context, name string) (int, error) {
	path := fmt.Sprintf("/admin/v1/groups/%s", name)

	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
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
		return 0, fmt.Errorf("decoding create group response: %w", err)
	}
	return result.ID, nil
}

// RenameGroup renames the group identified by id.
func (c *Client) RenameGroup(ctx context.Context, id int, name string) error {
	path := fmt.Sprintf("/admin/v1/groups/%d", id)
	body := struct {
		Name string `json:"name"`
	}{Name: name}

	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// DeleteGroups deletes groups by their IDs.
func (c *Client) DeleteGroups(ctx context.Context, ids []int) error {
	body := struct {
		GroupIDs []int `json:"groupIds"`
	}{GroupIDs: ids}

	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/groups", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// ---------------------------------------------------------------------------
// Group member operations
// ---------------------------------------------------------------------------

// groupMembersWire is the actual wire shape returned by the members endpoints.
type groupMembersWire struct {
	GroupID int `json:"groupId"`
	Clients []struct {
		MemberID   int    `json:"memberId"`
		MemberName string `json:"memberName"`
		MemberAlias string `json:"memberAlias"`
	} `json:"clients"`
	Devices []struct {
		MemberID   int    `json:"memberId"`
		MemberName string `json:"memberName"`
	} `json:"devices"`
}

func (w groupMembersWire) flatten() []GroupMember {
	var out []GroupMember
	for _, c := range w.Clients {
		out = append(out, GroupMember{GroupID: w.GroupID, MemberID: c.MemberID, MemberType: "client", MemberName: c.MemberName, MemberAlias: c.MemberAlias})
	}
	for _, d := range w.Devices {
		out = append(out, GroupMember{GroupID: w.GroupID, MemberID: d.MemberID, MemberType: "device", MemberName: d.MemberName})
	}
	return out
}

// ListMembers returns all group members across all groups.
func (c *Client) ListMembers(ctx context.Context, page, pageSize int) ([]GroupMember, error) {
	q := url.Values{}
	q.Set("Page", strconv.Itoa(page))
	q.Set("PageSize", strconv.Itoa(pageSize))
	path := "/admin/v1/groups/members?" + q.Encode()

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var envelope struct {
		Groups []groupMembersWire `json:"groups"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decoding members response: %w", err)
	}
	var out []GroupMember
	for _, g := range envelope.Groups {
		out = append(out, g.flatten()...)
	}
	return out, nil
}

// ListGroupMembers returns all members of a specific group.
func (c *Client) ListGroupMembers(ctx context.Context, groupID int) ([]GroupMember, error) {
	path := fmt.Sprintf("/admin/v1/groups/%d/members", groupID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var wire groupMembersWire
	if err := json.NewDecoder(resp.Body).Decode(&wire); err != nil {
		return nil, fmt.Errorf("decoding group members response: %w", err)
	}
	return wire.flatten(), nil
}

// AddMembers adds a member (client or device) to a group.
func (c *Client) AddMembers(ctx context.Context, groupID, memberID int, memberType string) error {
	body := struct {
		GroupID    int    `json:"groupId"`
		MemberID   int    `json:"memberId"`
		MemberType string `json:"memberType"`
	}{GroupID: groupID, MemberID: memberID, MemberType: memberType}

	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/groups/members", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// RemoveMembers removes a member from a group.
func (c *Client) RemoveMembers(ctx context.Context, groupID, memberID int) error {
	body := struct {
		GroupID  int `json:"groupId"`
		MemberID int `json:"memberId"`
	}{GroupID: groupID, MemberID: memberID}

	resp, err := c.doRequest(ctx, http.MethodDelete, "/admin/v1/groups/members", body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// ExportMembers downloads the raw bulk member export (e.g. CSV bytes).
func (c *Client) ExportMembers(ctx context.Context) ([]byte, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/groups/members/bulk", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var buf []byte
	tmp := make([]byte, 4096)
	for {
		n, readErr := resp.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if readErr != nil {
			break
		}
	}
	return buf, nil
}

// ImportMembers performs a bulk member import from structured records.
func (c *Client) ImportMembers(ctx context.Context, records []map[string]any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/groups/members/bulk", records)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// ---------------------------------------------------------------------------
// Group client operations
// ---------------------------------------------------------------------------

// AddClientsToGroup adds tracked clients to a group.
func (c *Client) AddClientsToGroup(ctx context.Context, groupID int, clientIDs []int) error {
	path := fmt.Sprintf("/admin/v1/groups/%d/clients", groupID)
	body := struct {
		ClientIDs []int `json:"clientIds"`
	}{ClientIDs: clientIDs}

	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// RemoveClientsFromGroup removes tracked clients from a group.
func (c *Client) RemoveClientsFromGroup(ctx context.Context, groupID int, clientIDs []int) error {
	path := fmt.Sprintf("/admin/v1/groups/%d/clients", groupID)
	body := struct {
		ClientIDs []int `json:"clientIds"`
	}{ClientIDs: clientIDs}

	resp, err := c.doRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// ---------------------------------------------------------------------------
// Group device operations
// ---------------------------------------------------------------------------

// AddDevicesToGroup adds devices to a group.
func (c *Client) AddDevicesToGroup(ctx context.Context, groupID int, deviceIDs []int) error {
	path := fmt.Sprintf("/admin/v1/groups/%d/devices", groupID)
	body := struct {
		DeviceIDs []int `json:"deviceIds"`
	}{DeviceIDs: deviceIDs}

	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetGroupMember retrieves a specific member from a group by type and member ID.
func (c *Client) GetGroupMember(ctx context.Context, groupID int, memberType string, memberID int) (*GroupMember, error) {
	path := fmt.Sprintf("/admin/v1/groups/%d/members/%s/%d", groupID, memberType, memberID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var m GroupMember
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("decoding group member: %w", err)
	}
	return &m, nil
}

// RemoveDevicesFromGroup removes devices from a group.
func (c *Client) RemoveDevicesFromGroup(ctx context.Context, groupID int, deviceIDs []int) error {
	path := fmt.Sprintf("/admin/v1/groups/%d/devices", groupID)
	body := struct {
		DeviceIDs []int `json:"deviceIds"`
	}{DeviceIDs: deviceIDs}

	resp, err := c.doRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
