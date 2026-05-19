// Package api provides the atadmin API client and shared types.
package api

// ErrorResponse represents an error payload returned by the ActivTrak API.
type ErrorResponse struct {
	Message string `json:"message"`
}

// ATClient represents a tracked user (client) in the ActivTrak system.
type ATClient struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	LogonDomain string `json:"logonDomain"`
	Alias       string `json:"alias"`
	Status      string `json:"status"`
	DeviceCount int    `json:"deviceCount"`
}

// DNTEntry represents a Do Not Track rule.
type DNTEntry struct {
	ID          int    `json:"id"`
	LogonDomain string `json:"logonDomain"`
	Username    string `json:"username"`
	IsGlobal    bool   `json:"isGlobal"`
}

// MergeUser describes a single source→target user merge record.
type MergeUser struct {
	SourceID int `json:"sourceId"`
	TargetID int `json:"targetId"`
}

// Group represents an ActivTrak user group.
type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
}

// GroupMember represents a member within a group.
type GroupMember struct {
	GroupID    int    `json:"groupId"`
	MemberID   int    `json:"memberId"`
	MemberType string `json:"memberType"`
}

// Consumer represents an ActivTrak account consumer (admin/report user).
type Consumer struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	UseSSO   bool   `json:"useSSO"`
}

// Device represents a tracked endpoint device in the ActivTrak system.
type Device struct {
	ID       int    `json:"id"`
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
}

// Signal represents an ActivTrak signal (alert rule).
type Signal struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// Condition represents a condition used within an alarm.
type Condition struct {
	ID       int    `json:"id"`
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Alarm represents an ActivTrak alarm.
type Alarm struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Enabled    bool        `json:"enabled"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// Schedule represents an ActivTrak work schedule.
type Schedule struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDefault bool   `json:"isDefault"`
}

// UserScheduleInfo describes a user's schedule assignment.
type UserScheduleInfo struct {
	UserID     int    `json:"userId"`
	ScheduleID int    `json:"scheduleId"`
	Name       string `json:"name"`
}

// ApiKey represents a Public API credential.
//
//nolint:revive // "api.ApiKey" stutter is intentional — matches the domain term.
type ApiKey struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	KeyPrefix  string `json:"keyPrefix"`
	CreatedAt  string `json:"createdAt"`
	LastUsedAt string `json:"lastUsedAt"`
}

// AuditLog represents an immutable administrative action record.
type AuditLog struct {
	ID           int    `json:"id"`
	Action       string `json:"action"`
	Actor        string `json:"actor"`
	Timestamp    string `json:"timestamp"`
	Details      string `json:"details"`
	AttachmentID string `json:"attachmentId"`
}
