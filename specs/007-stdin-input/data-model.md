# Data Model: Input via Stdin (`--from-stdin`)

No new persistent storage. No new API models.

## New Package: `internal/stdin`

```go
package stdin

import (
    "encoding/json"
    "fmt"
    "io"
)

// ReadJSON reads all bytes from r and unmarshals them into T.
// Returns an actionable error if r is empty or the JSON is invalid.
func ReadJSON[T any](r io.Reader) (T, error) {
    var zero T
    data, err := io.ReadAll(r)
    if err != nil {
        return zero, fmt.Errorf("--from-stdin: reading stdin: %w", err)
    }
    if len(data) == 0 {
        return zero, fmt.Errorf("--from-stdin: stdin is empty; pipe a JSON payload")
    }
    var result T
    if err := json.Unmarshal(data, &result); err != nil {
        return zero, fmt.Errorf("--from-stdin: invalid JSON: %w", err)
    }
    return result, nil
}

// ReadRecords reads all bytes from r and unmarshals them as a JSON array
// of objects. Compatible with the []map[string]any format produced by
// bulk.ParseFile for JSON files.
func ReadRecords(r io.Reader) ([]map[string]any, error) {
    data, err := io.ReadAll(r)
    if err != nil {
        return nil, fmt.Errorf("--from-stdin: reading stdin: %w", err)
    }
    if len(data) == 0 {
        return nil, fmt.Errorf("--from-stdin: stdin is empty; pipe a JSON array")
    }
    var records []map[string]any
    if err := json.Unmarshal(data, &records); err != nil {
        return nil, fmt.Errorf("--from-stdin: invalid JSON: %w", err)
    }
    return records, nil
}
```

## Existing Types Used as Stdin Payload Targets

### Group B — Typed struct commands

| Command | Payload Type | JSON Fields |
|---|---|---|
| `users update` | `api.UpdateUserRequest` | `displayName`, `firstName`, `lastName`, `timezone`, `tracked` |
| `users bulk *` | `api.BulkActionRequest` | `actions` ([]string), `data` ([]{entityId, revision}) |

**Note on `users update`**: `UpdateUserRequest` uses pointer fields (`*string`, `*bool`). A JSON field present in stdin will be set; an absent field is left nil and skipped by the PATCH call — identical behavior to the flag-based path.

**Note on `users bulk *`**: When `--from-stdin` is used, the caller must supply the full `BulkActionRequest` including the `actions` array (even though the subcommand name already implies the action). This is the most explicit and machine-friendly format. Alternative: accept just `{"data": [...]}` and inject the action name from the command — but this adds hidden behavior. Full payload is preferred.

### Group C — Confirmation bypass commands

| Command | Stdin Used? | Payload (if any) |
|---|---|---|
| `users delete` | No | None — `--from-stdin` just bypasses the y/N prompt |
| `consumers password set` | Yes | `{"password": "..."}` — replaces interactive prompt |

## `PasswordStdinPayload` (consumers password set)

```go
// PasswordStdinPayload is the JSON shape accepted by `consumers password set --from-stdin`.
// Defined inline in the command, not as a shared API model.
type passwordStdinPayload struct {
    Password string `json:"password"`
}
```

Validation: if `password` is empty after unmarshal, return `--from-stdin: "password" field is required`.

## Group A — Record-map commands (existing `--file` commands)

These commands already accept `[]map[string]any` from `bulk.ParseFile`. `stdin.ReadRecords` produces the same type. No type changes needed.
