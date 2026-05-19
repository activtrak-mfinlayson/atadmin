# Data Model: ActivTrak Admin API CLI Wrapper

Source: `docs/admin-swagger.json` — 110 endpoints, 11 resource groups.

---

## Client

Represents a tracked end-user (employee or workstation login identity).

| Field | Type | Notes |
|---|---|---|
| `id` | int | Numeric client ID; used in path parameters |
| `username` | string | Logon username (domain\user or UPN); also used as path key |
| `alias` | string | Display name override; empty string means no alias |
| `logon_domain` | string | Windows domain or empty |
| `status` | string | Active, Inactive, or DoNotTrack |
| `device_count` | int | Number of associated devices |
| `group_ids` | []int | Group memberships |

**Do Not Track (DNT) sub-entity**:

| Field | Type | Notes |
|---|---|---|
| `id` | int | DNT entry ID |
| `logon_domain` | string | Domain for the DNT rule |
| `username` | string | Username for the DNT rule |
| `is_global` | bool | If true, applies across all logon domains |

---

## Consumer

Represents an ActivTrak admin user (someone who logs into the ActivTrak dashboard).

| Field | Type | Notes |
|---|---|---|
| `id` | int | Numeric consumer ID |
| `username` | string | Email or login username |
| `role` | string | `admin`, `viewer`, `configurator`, `superadmin` |
| `use_sso` | bool | Whether SSO is used for this consumer |
| `viewable_group_ids` | []int | Groups this consumer can see in reports |

---

## Device

Represents a monitored computer with an installed ActivTrak agent.

| Field | Type | Notes |
|---|---|---|
| `id` | int | Numeric device ID |
| `hostname` | string | Computer hostname |
| `agent_status` | string | `active`, `inactive`, `uninstalled` |
| `last_seen` | datetime | Last agent check-in timestamp |
| `group_ids` | []int | Group memberships |

---

## Group

A named collection of Clients and/or Devices for reporting and policy scoping.

| Field | Type | Notes |
|---|---|---|
| `id` | int | Numeric group ID |
| `name` | string | Unique display name |
| `member_count` | int | Total members across all types |
| `client_ids` | []int | Client members |
| `device_ids` | []int | Device members |

**GroupMember sub-entity**:

| Field | Type | Notes |
|---|---|---|
| `member_id` | int | ID of the client or device |
| `member_type` | string | `client` or `device` |
| `group_id` | int | Group this membership belongs to |

---

## Alarm

A threshold-based alert that fires when activity conditions are met.

| Field | Type | Notes |
|---|---|---|
| `id` | int | Numeric alarm ID |
| `name` | string | Display name |
| `type` | string | Alarm type (e.g., `productivity`, `website`, `application`) |
| `conditions` | []Condition | One or more threshold conditions |
| `channels` | []Channel | Notification destinations (email, webhook) |
| `enabled` | bool | Whether the alarm is active |

**Condition sub-entity**: `field`, `operator`, `value` — e.g., `{field: "url", operator: "contains", value: "reddit.com"}`

**Channel sub-entity**: `type` (`email` or `webhook`), `target` (address or URL)

---

## Signal

A configurable notification definition (event-driven, as opposed to threshold-based Alarms).

| Field | Type | Notes |
|---|---|---|
| `id` | int | Numeric signal ID |
| `name` | string | Display name |
| `type` | string | Signal category |
| `enabled` | bool | Whether active |
| `channels` | []Channel | Same channel types as Alarm |

---

## Schedule

A time-based schedule (reporting or shift type) with assigned users.

| Field | Type | Notes |
|---|---|---|
| `id` | int | Numeric schedule ID |
| `name` | string | Display name |
| `type` | string | `reporting` or `shift` |
| `is_default` | bool | Whether this is the account-wide default |
| `user_ids` | []int | Users assigned to this schedule |

---

## ApiKey

A credential for external access to the ActivTrak Public API.

| Field | Type | Notes |
|---|---|---|
| `id` | int | Numeric key ID |
| `name` | string | Descriptive label |
| `key_prefix` | string | First characters of the key (never full value) |
| `created_at` | datetime | Creation timestamp |
| `last_used_at` | datetime | Last usage timestamp; null if never used |

---

## AuditLog

Immutable administrative action record.

| Field | Type | Notes |
|---|---|---|
| `id` | int | Entry ID |
| `action` | string | Action type (e.g., `UpdateUserPrivilege`) |
| `actor` | string | Consumer who performed the action |
| `timestamp` | datetime | When the action occurred |
| `details` | string | Free-text description |
| `attachment_id` | string | Optional attachment ID for downloadable detail |

Supports: `FromDate`, `ToDate`, `Filters`, `SortColumn`, `SortDescending`, `Page`, `PageSize`.

---

## Account Settings (composite)

Not a single entity — a collection of account-level configuration sub-resources. Each sub-resource has its own GET and SET shape.

| Sub-resource | Key fields |
|---|---|
| `privacy` | `show_activities`, `show_screenshots`, `show_alarms` |
| `sso` | `provider`, `entity_id`, `sso_url`, `enabled`, `eligible` |
| `role_access` | `[]{ resource, roles[] }` |
| `role_date_filter` | `[]{ filter, roles[] }` |
| `timezone` | `timezone_name`, `offset` |
| `agent_activity_duration` | `minutes` |
| `agent_audit` | `enabled`, `interval` |
| `passive_time` | `minutes`, `per_device` |
| `schedule_adherence` | `enabled`, `threshold_minutes` |
| `email_auto_detection` | `enabled` |
| `identity_new_agent_match` | `match_user` |
| `identity_search_active_threshold` | `days` |
| `license_approval_mode` | `mode` |
| `msp_license_overage` | `allowed`, `limit` |
| `local_timezone` | `show_local` |

---

## Named Profile (Config)

Not an API entity — a local config construct for multi-account support.

| Field | Type | Notes |
|---|---|---|
| `name` | string | Profile identifier; `default` if omitted |
| `token` | string | Bearer token for this account |
| `base_url` | string | API base URL for this account |
| `format` | string | Default output format (`table` or `json`) |
| `timeout` | duration | Request timeout (default `30s`) |

Config file location: `~/.config/atadmin/config.yaml`

```yaml
profiles:
  default:
    token: "..."
    base_url: "https://api.activtrak.com"
  staging:
    token: "..."
    base_url: "https://staging-api.activtrak.com"
```
