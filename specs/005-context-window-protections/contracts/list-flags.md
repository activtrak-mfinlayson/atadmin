# CLI Contract: Context Window Protection Flags

## Overview

Three flags are added to all list commands with `--json` output support. Together they form the context window protection contract for AI agent consumers.

---

## Flag: `--fields <keys>`

**Type**: string  
**Default**: `""` (disabled — full output)  
**Applied to**: All `list` commands with `--json`

### Behavior

- Comma-separated list of top-level JSON key names to retain in the output.
- Filtering is client-side: the full payload is fetched from the API, then stripped before writing to stdout.
- Keys that do not exist in the response objects are silently ignored (no error).
- Only top-level keys are supported in V1. Nested paths (e.g., `profile.email`) are out of scope.
- When combined with `--summary`, `--summary` wins and `--fields` is ignored.
- When `--json` is not passed, `--fields` is silently ignored.

### Examples

```sh
# Only return id and email for each user
atadmin users list --json --fields id,email

# MCP equivalent (auto-injects --json)
# Tool: users_list, params: {"fields": "id,email"}
```

### Schema (MCP tool parameter)

```json
{
  "name": "fields",
  "type": "string",
  "description": "Comma-separated top-level JSON keys to include (e.g. id,email)"
}
```

---

## Flag: `--summary`

**Type**: boolean  
**Default**: `false`  
**Applied to**: `list` commands returning arrays: `users`, `groups`, `clients`, `consumers`, `devices`, `alarms`, `auditlog`, `agents`

### Behavior

- Returns aggregate statistics instead of the full array.
- Output shape:
  ```json
  {
    "returned_items": 50,
    "total_items": 1234,
    "has_more": true
  }
  ```
- `total_items` is omitted when the underlying API response does not include a total count.
- `has_more` is `true` when there are additional pages (derived from `NextCursor` for cursor-based commands, or `returned_items == page_size` for offset-based commands where applicable).
- Short-circuits before `--fields`; combining both flags returns summary only.
- Active only when `--json` is also passed; otherwise ignored.

### Examples

```sh
# Find out how many users match a filter without loading all of them
atadmin users list --json --summary --filter active30days

# MCP equivalent
# Tool: users_list, params: {"summary": true, "filter": "active30days"}
```

### Schema (MCP tool parameter)

```json
{
  "name": "summary",
  "type": "boolean",
  "description": "Return aggregate statistics instead of full results"
}
```

---

## Behavior: Safe JSON Pagination

**Type**: automatic (no new flag)  
**Trigger**: `--json` is passed without an explicit `--limit` or `--page-size`

### Behavior

- When `--json` is set and the user has not explicitly provided a limit, the CLI defaults to 50 items.
- The default is applied by checking `cmd.Flags().Changed("limit")` (or `"page-size"`) before the API call.
- This prevents commands like `atadmin users list --json` from dumping thousands of items by default.
- Explicit values always take precedence: `atadmin users list --json --limit 200` returns up to 200 items.

### Commands affected

| Command | Flag name | Old default | New default when --json |
|---------|-----------|-------------|------------------------|
| `users list` | `--limit` | `0` (server default) | `50` |
| `agents list` | `--limit` | `0` (server default) | `50` |
| `auditlog list` | `--page-size` | `0` (server default) | `50` |
| `groups list` | `--page-size` | `50` | unchanged (already 50) |
| `clients list` | `--page-size` | `50` | unchanged (already 50) |
| `consumers list` | `--page-size` | `50` | unchanged (already 50) |
| `devices list` | `--page-size` | `50` | unchanged (already 50) |
| `alarms list` | `--page-size` | `50` | unchanged (already 50) |
| `signals list` | N/A | N/A | no change |
| `schedules list` | N/A | N/A | no change |
| `apikeys list` | N/A | N/A | no change |

### MCP note

The MCP `makeHandler` auto-injects `--json` for all tools where `HasJSONFlag` is true. This means safe pagination applies to every MCP tool call by default, protecting agent context windows without any extra configuration.
