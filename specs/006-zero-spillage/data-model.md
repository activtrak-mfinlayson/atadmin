# Data Model: Zero-Spillage Rule (Feature 006)

This feature adds no new persistence layer. It adds two pure-function helpers to `internal/output/` and reclassifies ~25 existing `fmt.Fprint*` call sites.

---

## New Types

### `JSONError` (internal/output — write-only, not stored)

The structured error emitted to stdout when `--json` is active and a command fails.

```go
type JSONError struct {
    Error      string `json:"error"`
    Suggestion string `json:"suggestion,omitempty"`
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `error` | string | yes | Human-readable error message (same text as the non-JSON error) |
| `suggestion` | string | no | Actionable next step for the user or agent to self-correct |

Serialized as a JSON object on a single logical block (indented with 2 spaces, newline-terminated). The `suggestion` key is omitted when empty (`omitempty`).

**Example — auth failure:**
```json
{
  "error": "loading profile \"default\": token not configured",
  "suggestion": "Run 'atadmin auth login' to configure credentials."
}
```

**Example — generic failure:**
```json
{
  "error": "users get: 404 Not Found"
}
```

---

## New Functions in `internal/output`

### `WriteError(out io.Writer, err error, suggestion string, asJSON bool)`

Writes the error to the correct stream in the correct format.

- If `asJSON` is true: marshals a `JSONError` to `out` (stdout).
- If `asJSON` is false: writes `"Error: <message>\n"` to `out` (stderr).

The `out` parameter is always passed by the caller — the function never reads `os.Stdout` or `os.Stderr` directly, keeping it fully testable.

```go
func WriteError(out io.Writer, err error, suggestion string, asJSON bool) {
    if asJSON {
        je := JSONError{Error: err.Error(), Suggestion: suggestion}
        enc := json.NewEncoder(out)
        enc.SetIndent("", "  ")
        _ = enc.Encode(je)
        return
    }
    _, _ = fmt.Fprintf(out, "Error: %s\n", err)
}
```

### `DetectJSONMode(args []string) bool`

Scans a slice of command-line arguments for any of the patterns that signal JSON output mode. Returns true if any match.

Patterns recognized:
- `--json`
- `--format json`
- `--format=json`
- `-f json`
- `-f=json`

```go
func DetectJSONMode(args []string) bool
```

---

## Call-site Reclassification

The following existing `fmt.Fprint*` call-sites in `internal/cmd/` must be changed from `cmd.OutOrStdout()` to `cmd.ErrOrStderr()`:

| File | Line (approx) | Current message | Action |
|---|---|---|---|
| users.go | ~229 | `"Updated user %d\n"` | → ErrOrStderr |
| users.go | ~269 | `"Delete user %d? [y/N] "` | → ErrOrStderr |
| users.go | ~274 | `"Aborted.\n"` | → ErrOrStderr |
| users.go | ~291 | `"Deleted user %d\n"` | → ErrOrStderr |
| users.go | ~364 | `"Added group(s) to user %d\n"` | → ErrOrStderr |
| users.go | ~419 | `"Removed group(s) from user %d\n"` | → ErrOrStderr |
| groups.go | ~218 | `"Renamed group %d to %q\n"` | → ErrOrStderr |
| groups.go | ~249 | `"Deleted %d groups\n"` | → ErrOrStderr |
| groups.go | ~251 | `"deleted"` | → ErrOrStderr |
| groups.go | ~371 | `"Added member %d (%s) to group %d\n"` | → ErrOrStderr |
| groups.go | ~405 | `"Removed member %d from group %d\n"` | → ErrOrStderr |
| groups.go | ~469 | `"Imported %d member records\n"` | → ErrOrStderr |
| groups.go | ~525 | `"Added %d clients to group %d\n"` | → ErrOrStderr |
| groups.go | ~561 | `"Removed %d clients from group %d\n"` | → ErrOrStderr |
| groups.go | ~617 | `"Added %d devices to group %d\n"` | → ErrOrStderr |
| groups.go | ~653 | `"Removed %d devices from group %d\n"` | → ErrOrStderr |
| consumers.go | ~134 | `"Created %d consumers\n"` | → ErrOrStderr |
| consumers.go | ~165 | `"Updated %d consumers\n"` | → ErrOrStderr |
| consumers.go | ~196 | `"Deleted %d consumers\n"` | → ErrOrStderr |
| consumers.go | ~198 | `"deleted"` | → ErrOrStderr |
| consumers.go | ~230 | `"Deleted %d consumers\n"` | → ErrOrStderr |

**Stay on OutOrStdout (data outputs):**
- `accounts.go`: `"OK"`, `strconv.FormatBool(v)`, URL strings — these are data values
- `signals.go:85`: signal ID integer — data (created resource ID)
- `signals.go:136`: signal ID integer in delete-with-json path — data
- `groups.go:191`: group ID — data
- `apikeys.go:85`: API key ID — data

---

## Modified Call-site in `root.go`

`Execute()` changes from:
```go
if err := root.Execute(); err != nil {
    _, _ = fmt.Fprintf(root.ErrOrStderr(), "Error: %s\nRun 'atadmin --help' for usage.\n", err)
    return err
}
```

To:
```go
asJSON := output.DetectJSONMode(os.Args[1:])
if err := root.Execute(); err != nil {
    out := root.ErrOrStderr()
    if asJSON {
        out = root.OutOrStdout()
    }
    output.WriteError(out, err, output.SuggestionFor(err), asJSON)
    return err
}
```

Note: the `"Run 'atadmin --help' for usage."` suffix is dropped in JSON mode (it's not machine-friendly) but could be added as a suggestion value instead.
