# Research: Zero-Spillage Rule (Feature 006)

## 1. Auditing stdout vs stderr classification

### Decision
Messages that are **data** stay on `cmd.OutOrStdout()`; messages that are **diagnostic confirmations** move to `cmd.ErrOrStderr()` unconditionally (not just in JSON mode).

### Rationale
A diagnostic message like "Updated user 123" is not data — it is a side-channel confirmation of a side effect. It will corrupt any downstream pipe regardless of whether `--json` is explicitly passed, because any tool parsing stdout for IDs or JSON objects will encounter unexpected text. Moving all confirmation messages to stderr fixes both the JSON case and the plain-text piped case.

### Data vs Diagnostic classification

| Pattern | Classification | Correct stream |
|---|---|---|
| `"Updated user %d\n"` | Diagnostic | ErrOrStderr |
| `"Deleted user %d\n"` | Diagnostic | ErrOrStderr |
| `"Aborted.\n"` | Diagnostic | ErrOrStderr |
| `"Added group(s) to user %d\n"` | Diagnostic | ErrOrStderr |
| `"Removed group(s) from user %d\n"` | Diagnostic | ErrOrStderr |
| `"Renamed group %d to %q\n"` | Diagnostic | ErrOrStderr |
| `"Added member ... to group %d\n"` | Diagnostic | ErrOrStderr |
| `"Removed member ... from group %d\n"` | Diagnostic | ErrOrStderr |
| `"Created %d consumers\n"` | Diagnostic | ErrOrStderr |
| `"Updated %d consumers\n"` | Diagnostic | ErrOrStderr |
| `"Deleted %d consumers\n"` | Diagnostic | ErrOrStderr |
| `"Delete user %d? [y/N] "` | Interactive prompt | ErrOrStderr |
| `strconv.Itoa(id)` (created signal/group ID) | Data (scalar) | OutOrStdout ✓ |
| `strconv.FormatBool(v)` (setting value) | Data (scalar) | OutOrStdout ✓ |
| URL strings (setting values) | Data (scalar) | OutOrStdout ✓ |
| `"OK"` (ping result) | Data (scalar status) | OutOrStdout ✓ |

### Alternatives considered
- **Only reclassify in JSON mode**: Would require threading a "json active" boolean down to every call site. More invasive and easy to miss. Rejected.
- **Leave confirmations on stdout, suppress in JSON mode**: Would require per-command JSON detection. More complex than the simple correct fix. Rejected.

---

## 2. JSON mode detection in Execute()

### Decision
Scan `os.Args` directly in `Execute()` for `--json`, `--format json`, and `--format=json` before calling `root.Execute()`. Store the result in a local bool and pass it to `output.WriteError()`.

### Rationale
By the time `root.Execute()` returns an error, the Cobra `PersistentPreRunE` may or may not have run (e.g., it won't run for flag parse errors). Reading `os.Args` directly is simple, zero-overhead, and reliable across all error paths including early failures.

### Implementation
```go
func detectJSONMode(args []string) bool {
    for i, arg := range args {
        if arg == "--json" {
            return true
        }
        if arg == "--format=json" || arg == "-f=json" {
            return true
        }
        if (arg == "--format" || arg == "-f") && i+1 < len(args) && args[i+1] == "json" {
            return true
        }
    }
    return false
}
```

This function lives in `internal/output/` alongside `WriteError()` so tests can exercise it directly.

### Alternatives considered
- **Global var set in PersistentPreRunE**: Simpler but misses flag-parse errors and breaks test isolation. Rejected.
- **Re-parse flags in Execute() using pflag**: Correct but heavyweight; adds a dependency. Rejected.

---

## 3. Suggestion text for structured errors

### Decision
Derive a suggestion string from the error type/message at the `Execute()` level using a small `SuggestionFor(err error) string` helper. Do not add suggestion logic inside individual commands.

### Rationale
Centralizing suggestion derivation in one place makes it easy to add new patterns without touching every command. The most common actionable suggestions map to well-known error patterns already present in the codebase.

### Suggestion mapping
| Error signal | Suggestion |
|---|---|
| HTTP 401 / "unauthorized" | `"Run 'atadmin auth login' to authenticate."` |
| HTTP 404 / "not found" | `"Check the resource ID and try again."` |
| HTTP 5xx / "server error" | `"The ActivTrak API encountered an error. Try again later."` |
| "loading profile" | `"Run 'atadmin auth login' to configure credentials."` |
| All other errors | `""` (empty — no suggestion) |

The helper inspects `err.Error()` for these substrings. It does not unwrap to a typed error to keep the dependency surface minimal.

### Alternatives considered
- **Wrap errors with suggestion at command level**: More precise but requires every command to know what suggestion text to add. Rejected.
- **Typed error with suggestion field**: Cleaner but requires API layer changes. Deferred to a later feature.

---

## 4. Interaction with existing `--json` per-command flags

### Finding
The codebase has two parallel JSON mode signals:
1. Global `--format/-f json` on the root command
2. Per-command `--json` boolean flags on individual read commands

Both should trigger JSON error output. The `DetectJSONMode(os.Args)` scanner handles both patterns. For mutation commands that do not have a `--json` flag, only `--format json` applies.

### Decision
No changes to per-command `--json` flags. The scanner covers both. No unification pass needed for this feature.
