# CLI Command Contracts: Identity API

**Feature**: 003-identity-api-cli  
**Date**: 2026-05-19

This file is the authoritative contract for every new command and flag introduced by this feature. Tests are written against these contracts.

---

## `atadmin users`

Top-level command group for identity entity management.

```
atadmin users [subcommand]
```

Subcommands: `list`, `get`, `update`, `delete`, `groups`, `bulk`

---

## `atadmin users list`

List identity entities with optional filtering, searching, and pagination.

```
atadmin users list [flags]
```

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--filter` | string | — | Filter type: `tracked`, `untracked`, `active30days`, `groupcount`, `nogroupcount`, `nodisplayname`, `orphans`, `unlicensed`, `activeUsers`, `inactiveUsers`, `pendingLicense` |
| `--search` | string | — | Search term to match against identities |
| `--search-type` | string | — | Field to search: `email`, `upn`, `employeeid`, `displayname`, `alias`, `all`, etc. |
| `--sort` | string | — | Sort field (e.g., `displayname`, `lastlog`, `groupcount`) |
| `--sort-dir` | string | — | Sort direction: `asc` or `desc` |
| `--limit` | int | 0 (server default) | Maximum number of results to return |
| `--cursor` | string | — | Pagination cursor from a previous response |
| `--json` | bool | false | Output raw JSON instead of table |

### Table Output Columns

```
ID    DISPLAY NAME    STATUS       GROUPS    TRACKED
```

### JSON Output

Raw `UsersPage` JSON: `{"results":[...],"cursor":"...","totalCount":N}`

### Exit Codes

| Code | Condition |
|---|---|
| 0 | Success (including empty result) |
| 1 | API error, network error, invalid flag |

---

## `atadmin users get <id>`

Retrieve full details for a single identity entity.

```
atadmin users get <id> [flags]
```

### Arguments

| Argument | Type | Required | Description |
|---|---|---|---|
| `id` | int64 | yes | Entity ID |

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--json` | bool | false | Output raw JSON |

### Human-Readable Output

Key-value display:

```
id:               12345
revision:         3
displayName:      Alice Smith
firstName:        Alice
lastName:         Smith
emails:           alice@example.com
upns:             alice@corp.example.com
groups:           Engineering (42), All Users (1)
primaryGroup:     Engineering
tracked:          true
status:           active
timezone:         America/Chicago
agents:           ALICE-PC (approved), ALICE-LAPTOP (approved)
created:          2024-01-15T10:30:00Z
updated:          2024-01-20T14:25:30Z
```

### JSON Output

Raw `IdentityDetailsResponse` JSON.

---

## `atadmin users update <id>`

Patch one or more scalar fields on an identity entity. Automatically fetches the current revision unless `--revision` is supplied.

```
atadmin users update <id> [flags]
```

### Arguments

| Argument | Type | Required | Description |
|---|---|---|---|
| `id` | int64 | yes | Entity ID |

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--display-name` | string | — | New display name |
| `--first-name` | string | — | New first name |
| `--last-name` | string | — | New last name |
| `--timezone` | string | — | New timezone (IANA, e.g. `America/Chicago`) |
| `--tracked` | bool | — | Set tracking state |
| `--revision` | int64 | 0 (auto-fetch) | Explicit revision; skips the auto-fetch GET |
| `--json` | bool | false | Output updated entity as JSON |

### Behavior

At least one of `--display-name`, `--first-name`, `--last-name`, `--timezone`, or `--tracked` must be provided. If none are provided, the command exits with an error.

On success in non-JSON mode: prints `Updated user <id>` to stdout.

---

## `atadmin users delete <id>`

Delete an identity entity. Prompts for confirmation in TTY mode.

```
atadmin users delete <id> [flags]
```

### Arguments

| Argument | Type | Required | Description |
|---|---|---|---|
| `id` | int64 | yes | Entity ID |

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--revision` | int64 | 0 (auto-fetch) | Explicit revision |
| `--yes` | bool | false | Skip confirmation prompt (required in non-TTY mode) |

### Behavior

In TTY mode without `--yes`: prompts `Delete user <id>? [y/N]`.  
In non-TTY mode without `--yes`: fails with an error directing the operator to pass `--yes`.  
On success: prints `Deleted user <id>` to stdout.

---

## `atadmin users groups add <userId> <groupId>`

Add one or more groups to an identity. Automatically fetches revision.

```
atadmin users groups add <userId> <groupId> [flags]
atadmin users groups add <userId> --group-ids 1,2,3 [flags]
```

### Arguments (single group form)

| Argument | Type | Required | Description |
|---|---|---|---|
| `userId` | int64 | yes | Entity ID |
| `groupId` | int | conditional | Group ID (mutually exclusive with `--group-ids`) |

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--group-ids` | string | — | Comma-separated group IDs (alternative to positional arg) |
| `--revision` | int64 | 0 (auto-fetch) | Explicit revision |

### Behavior

On success: prints `Added group(s) to user <userId>` to stdout.

---

## `atadmin users groups remove <userId> <groupId>`

Remove one or more groups from an identity.

```
atadmin users groups remove <userId> <groupId> [flags]
atadmin users groups remove <userId> --group-ids 1,2,3 [flags]
```

Same flag and behavior pattern as `groups add`.  
On success: prints `Removed group(s) from user <userId>` to stdout.

---

## `atadmin users bulk <action>`

Apply a bulk action to multiple identity entities by ID.

```
atadmin users bulk start-tracking --ids 1,2,3
atadmin users bulk stop-tracking  --ids 1,2,3
atadmin users bulk delete-entity  --ids 1,2,3
atadmin users bulk delete-data    --ids 1,2,3
```

### Subcommands

| Subcommand | API action | Description |
|---|---|---|
| `start-tracking` | `StartTracking` | Enable tracking for entities |
| `stop-tracking` | `StopTracking` | Disable tracking for entities |
| `delete-entity` | `DeleteEntity` | Permanently delete entity records |
| `delete-data` | `DeleteData` | Delete recorded activity data for entities |

### Flags (on each subcommand)

| Flag | Type | Required | Description |
|---|---|---|---|
| `--ids` | string | yes | Comma-separated entity IDs |
| `--json` | bool | false | Output `BulkActionResponse` JSON |

### Behavior

1. Parse `--ids` into a slice of int64.
2. Concurrently fetch revision for each ID (max 10 in-flight).
3. Report pre-flight errors (entity not found) to stderr; skip those IDs.
4. POST the bulk action request.
5. Print a summary table of successes and failures.
6. Exit code 0 if all succeeded; 1 if any failures.

### Human-Readable Output

```
ENTITY ID    RESULT
1            ok
2            ok
3            failed: revision conflict
```

---

## `atadmin agents list`

List agent (device) entities across the account.

```
atadmin agents list [flags]
```

Same flags as `atadmin users list`. Uses `GET /identity/v1/agents` which returns the same `UsersPage` shape.

### Table Output Columns

```
USER ID    USERNAME    DOMAIN    ALIAS    LICENSE    LAST LOG
```

### JSON Output

Raw `UsersPage` JSON (same shape as users list, `results` contains `IdentityDetailsResponse` with `agents` field populated).

---

## Error Conventions

All commands follow the existing project error conventions:

| Error Type | Output |
|---|---|
| 401 Unauthorized | `Error: unauthorized: your token may have expired. Try running 'atadmin auth login'` |
| 403 Forbidden | `Error: forbidden: your account role may not have permission for this operation` |
| 404 Not Found | `Error: not found: the requested resource does not exist` |
| 409 Conflict | `Error: conflict (409): the entity was modified concurrently. Re-fetch with 'atadmin users get <id>' and retry` |
| 5xx / timeout | `Error: server error (5xx): try again later` |

All error text goes to stderr. Exit code is always 1 on error.
