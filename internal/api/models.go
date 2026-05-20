// Package api provides the atadmin API client and shared types.
package api

// ErrorResponse represents an error payload returned by the ActivTrak API.
type ErrorResponse struct {
	Message string `json:"message"`
}

// ATClient represents a tracked user (client) in the ActivTrak system.
type ATClient struct {
	ID          int    `json:"id"`
	Username    string `json:"name"`
	LogonDomain string `json:"domain"`
	Alias       string `json:"alias"`
	Status      string `json:"status"`
	DeviceCount int    `json:"deviceCount"`
}

// DNTEntry represents a Do Not Track rule.
type DNTEntry struct {
	ID          int    `json:"id"`
	LogonDomain string `json:"logondomain"`
	Username    string `json:"username"`
	IsGlobal    bool   `json:"globaluser"`
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
	MemberName string `json:"memberName"`
	MemberAlias string `json:"memberAlias"`
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
	Hostname string `json:"name"`
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
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDefault bool   `json:"isDefault"`
}

// UserScheduleInfo describes a user's schedule assignment.
type UserScheduleInfo struct {
	UserID       string `json:"userId"`
	UserName     string `json:"userName"`
	ScheduleID   string `json:"scheduleId"`
	ScheduleName string `json:"scheduleName"`
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

// ---------------------------------------------------------------------------
// Identity API types
// ---------------------------------------------------------------------------

// IdentityField wraps a single scalar field value with its field-level ID and
// the source system that last set the value.
type IdentityField struct {
	Value  string `json:"value"`
	ID     string `json:"id"`
	Source string `json:"source"`
}

// IdentityAgent is a device/client associated with an Identity entity.
type IdentityAgent struct {
	UserID        int    `json:"userId"`
	UserName      string `json:"userName"`
	LogonDomain   string `json:"logonDomain"`
	Alias         string `json:"alias"`
	Tracked       bool   `json:"tracked"`
	LicenseStatus string `json:"licenseStatus"`
	LastLog       string `json:"lastLog"`
	FirstLog      string `json:"firstLog"`
	Deleted       bool   `json:"deleted"`
}

// IdentityGroupRef is a group membership record embedded in an Identity.
type IdentityGroupRef struct {
	GroupID       int    `json:"groupId"`
	GroupName     string `json:"groupName"`
	GroupType     int    `json:"groupType"`
	GroupTypeName string `json:"groupTypeName"`
}

// Identity represents a logical person (entity) in the ActivTrak Identity system.
type Identity struct {
	ID               int64              `json:"id"`
	Revision         int64              `json:"revision"`
	DisplayName      *IdentityField     `json:"displayName"`
	FirstName        *IdentityField     `json:"firstName"`
	MiddleName       *IdentityField     `json:"middleName"`
	LastName         *IdentityField     `json:"lastName"`
	Emails           []IdentityField    `json:"emails"`
	UPNs             []IdentityField    `json:"upns"`
	EmployeeIDs      []IdentityField    `json:"employeeIds"`
	Groups           []IdentityGroupRef `json:"groups"`
	PrimaryGroupID   *int               `json:"primaryGroupId"`
	PrimaryGroupName string             `json:"primaryGroupName"`
	Agents           []IdentityAgent    `json:"agents"`
	Tracked          bool               `json:"tracked"`
	Status           string             `json:"status"`
	Timezone         *IdentityField     `json:"timezone"`
	Created          string             `json:"created"`
	Updated          string             `json:"updated"`
	DisplayID        *IdentityField     `json:"displayId"`
}

// UsersPage is the paginated response for Identity list operations.
type UsersPage struct {
	Results    []Identity `json:"results"`
	Cursor     string     `json:"cursor"`
	TotalCount int        `json:"totalCount"`
}

// IdentityListParams are the query parameters for listing identities or agents.
type IdentityListParams struct {
	Search     string
	SearchType string
	Filters    []string
	Sort       string
	SortDir    string
	Limit      int    // 0 = omit (server default)
	Cursor     string
}

// UpdateUserRequest holds the fields to patch on a PATCH call.
// Only non-nil pointer fields are sent in the request body.
type UpdateUserRequest struct {
	DisplayName *string
	FirstName   *string
	LastName    *string
	Timezone    *string
	Tracked     *bool
}

// BulkActionRequest is the request body for /identity/v1/entities/bulk.
type BulkActionRequest struct {
	Actions []string         `json:"actions"`
	Data    []BulkEntityData `json:"data"`
}

// BulkEntityData is one entity entry within a bulk action.
type BulkEntityData struct {
	EntityID int `json:"entityId"`
	Revision int `json:"revision"`
}

// BulkActionResponse is the response from /identity/v1/entities/bulk.
type BulkActionResponse struct {
	Successful []BulkEntitySuccess `json:"successful"`
	Failures   []BulkEntityFailure `json:"failures"`
}

// BulkEntitySuccess records a successful bulk action outcome for one entity.
type BulkEntitySuccess struct {
	EntityID int `json:"entityId"`
}

// BulkEntityFailure records a failed bulk action outcome for one entity.
type BulkEntityFailure struct {
	EntityID int    `json:"entityId"`
	Message  string `json:"message"`
}
