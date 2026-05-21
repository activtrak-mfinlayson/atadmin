# Data Model: Context Window Protections

## New Types (internal/output package)

### SummaryResult

Returned by `--summary` flag on any list command.

```go
// SummaryResult is the JSON shape returned by --summary on list commands.
type SummaryResult struct {
    ReturnedItems int  `json:"returned_items"`          // always present
    TotalItems    *int `json:"total_items,omitempty"`   // omitted when API doesn't provide total
    HasMore       bool `json:"has_more"`                // true when pagination has more pages
}
```

**Example output (cursor-based, more pages exist)**:
```json
{
  "returned_items": 50,
  "has_more": true
}
```

**Example output (offset-based, last page)**:
```json
{
  "returned_items": 23,
  "has_more": false
}
```

## New Functions (internal/output package)

### FilterFields

```go
// FilterFields returns data with only the specified top-level keys retained.
// Handles map[string]any (single object) and slices thereof (array of objects).
// Other types are returned unchanged.
func FilterFields(data any, fields []string) any
```

**Input/output examples**:

Input (array of objects):
```json
[{"id": 1, "email": "a@b.com", "status": "active", "groups": [...]}, ...]
```
With `fields = ["id", "email"]`:
```json
[{"id": 1, "email": "a@b.com"}, ...]
```

Input (single object):
```json
{"id": 1, "email": "a@b.com", "status": "active"}
```
With `fields = ["id"]`:
```json
{"id": 1}
```

### JSONSummary

```go
// JSONSummary writes a SummaryResult as indented JSON to out.
func JSONSummary(out io.Writer, returned int, total *int, hasMore bool) error
```

## Flag Additions to Cobra Commands

No new structs at the Cobra layer — the new flags bind to primitive Go types:

| Flag | Type | Commands |
|------|------|---------|
| `--fields` | `string` | All list commands with `--json` |
| `--summary` | `bool` | List commands returning paginated arrays |

## Validation Rules

- `--fields` with an empty string is treated as "no filtering" (passthrough).
- `--fields` with keys that don't exist in the response silently produces objects with only the keys that did match (standard map lookup behavior; no error).
- `--summary` and `--fields` are mutually exclusive at output time: `--summary` wins and `--fields` is ignored.
- `--summary` without `--json` is silently ignored (summary output is always JSON; the flag is a no-op in table mode).

## State Transitions

Not applicable. This feature is stateless output transformation only.
