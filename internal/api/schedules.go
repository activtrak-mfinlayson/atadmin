package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// scheduleWire is the wrapper shape returned by GET /admin/v1/schedules.
// Each list item has a "schedule" key.
type scheduleWire struct {
	Schedule Schedule `json:"schedule"`
}

// userScheduleWire is the shape returned by GET /admin/v1/schedules/*/users.
type userScheduleWire struct {
	ScheduleInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"scheduleInfo"`
	UserInfo struct {
		UserID   string `json:"userId"`
		UserName string `json:"userName"`
	} `json:"userInfo"`
}

// ListSchedules returns all schedules for the account.
// GET /admin/v1/schedules
func (c *Client) ListSchedules(ctx context.Context) ([]Schedule, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/admin/v1/schedules", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var raw []scheduleWire
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding schedules response: %w", err)
	}
	out := make([]Schedule, len(raw))
	for i, r := range raw {
		out[i] = r.Schedule
	}
	return out, nil
}

// GetSchedule returns a single schedule by UUID.
// GET /admin/v1/schedules/{id}
func (c *Client) GetSchedule(ctx context.Context, id string) (*Schedule, error) {
	path := fmt.Sprintf("/admin/v1/schedules/%s", id)
	return c.getSchedule(ctx, path)
}

// CreateSchedule creates a new schedule and returns its assigned UUID.
// POST /admin/v1/schedule
func (c *Client) CreateSchedule(ctx context.Context, body map[string]any) (string, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/admin/v1/schedule", body)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return "", err
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding create schedule response: %w", err)
	}
	return result.ID, nil
}

// DeleteSchedule removes the schedule with the given UUID.
// DELETE /admin/v1/schedules/{id}
func (c *Client) DeleteSchedule(ctx context.Context, id string) error {
	path := fmt.Sprintf("/admin/v1/schedules/%s", id)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetReportingDefault returns the default reporting schedule.
// GET /admin/v1/schedules/reporting/default
func (c *Client) GetReportingDefault(ctx context.Context) (*Schedule, error) {
	return c.getSchedule(ctx, "/admin/v1/schedules/reporting/default")
}

// SetReportingDefault sets the default reporting schedule to the given scheduleID.
// PUT /admin/v1/schedules/reporting/default/{scheduleId}
func (c *Client) SetReportingDefault(ctx context.Context, scheduleID string) error {
	path := fmt.Sprintf("/admin/v1/schedules/reporting/default/%s", scheduleID)
	resp, err := c.doRequest(ctx, http.MethodPut, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetShiftDefault returns the default shift schedule.
// GET /admin/v1/schedules/shift/default
func (c *Client) GetShiftDefault(ctx context.Context) (*Schedule, error) {
	return c.getSchedule(ctx, "/admin/v1/schedules/shift/default")
}

// SetShiftDefault sets the default shift schedule to the given scheduleID.
// PUT /admin/v1/schedules/shift/default/{scheduleId}
func (c *Client) SetShiftDefault(ctx context.Context, scheduleID string) error {
	path := fmt.Sprintf("/admin/v1/schedules/shift/default/%s", scheduleID)
	resp, err := c.doRequest(ctx, http.MethodPut, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetReportingUsers returns users assigned to the reporting schedule.
// GET /admin/v1/schedules/reporting/users
func (c *Client) GetReportingUsers(ctx context.Context) ([]UserScheduleInfo, error) {
	return c.getScheduleUsers(ctx, "/admin/v1/schedules/reporting/users")
}

// RemoveReportingUsers removes the given user IDs from the reporting schedule.
// DELETE /admin/v1/schedules/reporting/users
func (c *Client) RemoveReportingUsers(ctx context.Context, ids []int) error {
	return c.deleteScheduleUsers(ctx, "/admin/v1/schedules/reporting/users", ids)
}

// GetShiftUsers returns users assigned to the shift schedule.
// GET /admin/v1/schedules/shift/users
func (c *Client) GetShiftUsers(ctx context.Context) ([]UserScheduleInfo, error) {
	return c.getScheduleUsers(ctx, "/admin/v1/schedules/shift/users")
}

// RemoveShiftUsers removes the given user IDs from the shift schedule.
// DELETE /admin/v1/schedules/shift/users
func (c *Client) RemoveShiftUsers(ctx context.Context, ids []int) error {
	return c.deleteScheduleUsers(ctx, "/admin/v1/schedules/shift/users", ids)
}

// GetScheduleUsers returns the users assigned to the schedule with the given UUID.
// GET /admin/v1/schedules/{scheduleId}/users
func (c *Client) GetScheduleUsers(ctx context.Context, scheduleID string) ([]UserScheduleInfo, error) {
	path := fmt.Sprintf("/admin/v1/schedules/%s/users", scheduleID)
	return c.getScheduleUsers(ctx, path)
}

// SetScheduleUsers assigns the given user IDs to the schedule with the given UUID.
// PUT /admin/v1/schedules/{scheduleId}/users
func (c *Client) SetScheduleUsers(ctx context.Context, scheduleID string, userIDs []int) error {
	path := fmt.Sprintf("/admin/v1/schedules/%s/users", scheduleID)
	body := struct {
		UserIDs []int `json:"userIds"`
	}{UserIDs: userIDs}
	resp, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// MoveUserToSchedule moves a single user to the given schedule.
// PUT /admin/v1/schedules/{scheduleId}/user/{userId}
func (c *Client) MoveUserToSchedule(ctx context.Context, scheduleID string, userID int) error {
	path := fmt.Sprintf("/admin/v1/schedules/%s/user/%d", scheduleID, userID)
	resp, err := c.doRequest(ctx, http.MethodPut, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// GetUserReportingSchedule returns the reporting schedule assigned to a user.
// GET /admin/v1/user/{userId}/schedule/reporting
func (c *Client) GetUserReportingSchedule(ctx context.Context, userID int) (*Schedule, error) {
	path := fmt.Sprintf("/admin/v1/user/%d/schedule/reporting", userID)
	return c.getSchedule(ctx, path)
}

// GetUserShiftSchedule returns the shift schedule assigned to a user.
// GET /admin/v1/user/{userId}/schedule/shift
func (c *Client) GetUserShiftSchedule(ctx context.Context, userID int) (*Schedule, error) {
	path := fmt.Sprintf("/admin/v1/user/%d/schedule/shift", userID)
	return c.getSchedule(ctx, path)
}

// RemoveUserFromReportingSchedules removes a user from all reporting schedule assignments.
// DELETE /admin/v1/user/{userId}/schedule/reporting
func (c *Client) RemoveUserFromReportingSchedules(ctx context.Context, userID int) error {
	path := fmt.Sprintf("/admin/v1/user/%d/schedule/reporting", userID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// RemoveUserFromShiftSchedules removes a user from all shift schedule assignments.
// DELETE /admin/v1/user/{userId}/schedule/shift
func (c *Client) RemoveUserFromShiftSchedules(ctx context.Context, userID int) error {
	path := fmt.Sprintf("/admin/v1/user/%d/schedule/shift", userID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}

// ---------------------------------------------------------------------------
// internal helpers
// ---------------------------------------------------------------------------

func (c *Client) getSchedule(ctx context.Context, path string) (*Schedule, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var schedule Schedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("decoding schedule from %s: %w", path, err)
	}
	return &schedule, nil
}

func (c *Client) getScheduleUsers(ctx context.Context, path string) ([]UserScheduleInfo, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var raw []userScheduleWire
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding schedule users from %s: %w", path, err)
	}
	out := make([]UserScheduleInfo, len(raw))
	for i, r := range raw {
		out[i] = UserScheduleInfo{
			UserID:       r.UserInfo.UserID,
			UserName:     r.UserInfo.UserName,
			ScheduleID:   r.ScheduleInfo.ID,
			ScheduleName: r.ScheduleInfo.Name,
		}
	}
	return out, nil
}

func (c *Client) deleteScheduleUsers(ctx context.Context, path string, ids []int) error {
	body := struct {
		UserIDs []int `json:"userIds"`
	}{UserIDs: ids}
	resp, err := c.doRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkResponse(resp)
}
