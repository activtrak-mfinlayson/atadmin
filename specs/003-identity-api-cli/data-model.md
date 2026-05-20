# Data Model: Identity API CLI Commands

**Feature**: 003-identity-api-cli  
**Date**: 2026-05-19

## New Types (to add to `internal/api/models.go`)

### IdentityField

Wraps a single scalar field value. Many Identity entity fields use this shape instead of a bare string, because the API tracks the field-level ID and the source system that last set the value.

```go
type IdentityField struct {
    Value  string `json:"value"`
    ID     string `json:"id"`
    Source string `json:"source"`
}
```

Helper — returns the `.Value` string or empty string when the field pointer is nil:
```go
func fieldValue(f *IdentityField) string {
    if f == nil { return "" }
    return f.Value
}
```

### IdentityAgent

A device/client associated with an identity entity.

```go
type IdentityAgent struct {
    UserID        int    `json:"userId"`
    UserName      string `json:"userName"`
    LogonDomain   string `json:"logonDomain"`
    Alias         string `json:"alias"`
    Tracked       bool   `json:"tracked"`
    LicenseStatus string `json:"licenseStatus"` // "approved" | "pending" | "deleted"
    LastLog       string `json:"lastLog"`
    FirstLog      string `json:"firstLog"`
    Deleted       bool   `json:"deleted"`
}
```

### IdentityGroupRef

A group membership record embedded in an Identity response (distinct from the Admin API's `Group` type).

```go
type IdentityGroupRef struct {
    GroupID       int    `json:"groupId"`
    GroupName     string `json:"groupName"`
    GroupType     int    `json:"groupType"`
    GroupTypeName string `json:"groupTypeName"`
}
```

### Identity

The primary entity type. All pointer fields are optional/nullable in the API response.

```go
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
    Status           string             `json:"status"`  // "active" | "inactive" | "unlicensed"
    Timezone         *IdentityField     `json:"timezone"`
    Created          string             `json:"created"`
    Updated          string             `json:"updated"`
    DisplayID        *IdentityField     `json:"displayId"`
}
```

### UsersPage

Paginated list response for both `GET /identity/v1/entities` and `GET /identity/v1/agents`.

```go
type UsersPage struct {
    Results    []Identity `json:"results"`
    Cursor     string     `json:"cursor"`
    TotalCount int        `json:"totalCount"`
}
```

### IdentityListParams

Query parameter bundle for list operations. Zero values are omitted from the request.

```go
type IdentityListParams struct {
    Search     string   // --search
    SearchType string   // --search-type
    Filters    []string // --filter (repeatable)
    Sort       string   // --sort
    SortDir    string   // --sort-dir
    Limit      int      // --limit; 0 = omit (server default)
    Cursor     string   // --cursor (pagination token)
}
```

### UpdateUserRequest

The set of fields that `atadmin users update` may modify. Only non-nil fields are sent in the PATCH body.

```go
type UpdateUserRequest struct {
    DisplayName *string // sent as {"value":"..."}
    FirstName   *string
    LastName     *string
    Timezone     *string
    Tracked      *bool  // sent as bare boolean "tracking" key
}
```

### BulkActionRequest / BulkActionResponse

```go
type BulkActionRequest struct {
    Actions []string         `json:"actions"` // "StartTracking" | "StopTracking" | "DeleteData" | "DeleteEntity"
    Data    []BulkEntityData `json:"data"`
}

type BulkEntityData struct {
    EntityID int `json:"entityId"`
    Revision int `json:"revision"`
}

type BulkActionResponse struct {
    Successful []BulkEntitySuccess `json:"successful"`
    Failures   []BulkEntityFailure `json:"failures"`
}

type BulkEntitySuccess struct {
    EntityID int `json:"entityId"`
}

type BulkEntityFailure struct {
    EntityID int    `json:"entityId"`
    Message  string `json:"message"`
}
```

## Entity Relationships

```
Identity (1) ──has many──> IdentityAgent    (agents field)
Identity (1) ──belongs to──> IdentityGroupRef[] (groups field)
Identity (1) ──has many──> IdentityField    (emails, upns, employeeIds)
Identity (1) ──has one──> IdentityField     (displayName, firstName, lastName, timezone)
```

## State Transitions

```
Identity.tracked: true ←──────────────────→ false
  via: atadmin users update --tracked=false
       atadmin users bulk stop-tracking
       atadmin users bulk start-tracking

Identity.status: active | inactive | unlicensed  (read-only, derived from agent activity)

Identity existence: present → deleted
  via: atadmin users delete <id>
       atadmin users bulk delete-entity
```

## Optimistic Concurrency

Every identity has an `int64` revision counter. The counter increments on every successful mutation. The CLI read-modify-write flow:

```
GET /identity/v1/entities/{id}        → capture revision N
PATCH /identity/v1/entities/{id}/revision/N  → success → revision becomes N+1
                                              → 409 if revision has changed
```

The revision is embedded in the `Identity` struct and always returned by `GetUser`.

## Wire Format Notes

The API wraps scalar fields in `{"value":"...","id":"...","source":"..."}` objects. The `IdentityField` type maps this directly. When constructing PATCH request bodies (`FieldEditsDto`), scalar fields are sent as `{"value":"..."}` objects (the `id` and `source` fields are omitted on write).

The `agents` field in `IdentityDetailsResponse` uses camelCase keys (`userName`, `logonDomain`, `userId`, `lastLog`, `firstLog`, `licenseStatus`) that map directly to `IdentityAgent` JSON tags.
