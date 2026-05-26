# Research: Input via Stdin (`--from-stdin`)

## Decisions & Findings

---

### Decision 1: Scope — which commands get `--from-stdin`

**Decision**: Two groups of commands:

**Group A — file-based bulk commands** (currently accept `--file path.json`): receive `--from-stdin` as an alternative input source. Stdin is parsed as JSON using the same `[]map[string]any` format that `bulk.ParseFile` produces.

| Command | Current `--file` flag |
|---|---|
| `clients merge-bulk` | Yes |
| `clients unmerge-bulk` | Yes |
| `clients alias-bulk` | Yes |
| `clients donottrack add-bulk` | Yes |
| `clients donottrack remove-bulk` | Yes |
| `consumers delete-bulk` | Yes |
| `consumers chrome-users import` | Yes |
| `consumers create` | Yes |
| `consumers update` | Yes |
| `groups members import` | Yes |
| `hrdc import` | Yes |
| `schedules create` | Yes |
| `signals create` | Yes |
| `signals update` | Yes |
| `alarms create` | Yes |
| `alarms update` | Yes |

**Group B — typed-struct mutation commands** (currently use individual flags): receive `--from-stdin` as a full payload replacement. Stdin is parsed directly into the command's Go request struct.

| Command | Request Struct |
|---|---|
| `users update` | `api.UpdateUserRequest` |
| `users bulk start-tracking` | `api.BulkActionRequest` |
| `users bulk stop-tracking` | `api.BulkActionRequest` |
| `users bulk delete-entity` | `api.BulkActionRequest` |
| `users bulk delete-data` | `api.BulkActionRequest` |

**Group C — confirmation-prompt commands**: receive `--from-stdin` as a prompt-bypass signal. No payload is read from stdin; the flag just means "skip interactive confirmation."

| Command | Current prompt behavior |
|---|---|
| `users delete` | `tty.IsTerminal()` scanner prompt |
| `consumers password set` | `tty.IsTerminal()` masked password read |

**Rationale**: Only commands where the flag provides real ergonomic value are included. Simple one-arg setters (`clients alias-set <id> <alias>`, `consumers role-set <id> <role>`, etc.) are excluded — a single positional argument has no escaping complexity.

**Alternatives considered**: A global persistent flag on the root command was rejected because it would conflict with stdin usage in non-mutating commands and is harder to document per-command.

---

### Decision 2: Package design for stdin reading

**Decision**: New package `internal/stdin` with two functions:

```go
// ReadJSON unmarshals the full stdin content into T.
func ReadJSON[T any](r io.Reader) (T, error)

// ReadRecords parses stdin as a JSON array of objects.
// Compatible with the []map[string]any format produced by bulk.ParseFile.
func ReadRecords(r io.Reader) ([]map[string]any, error)
```

**Rationale**: Keeps the two use patterns (typed struct vs record-map) explicit. Generics (`ReadJSON[T]`) are idiomatic in Go 1.21+ and avoid interface casting at call sites. `ReadRecords` is a thin wrapper over `json.Unmarshal` into `[]map[string]any`, matching what `bulk.parseJSON` already does — this keeps Group A commands consistent with their existing `--file` behavior.

**Alternatives considered**: Adding methods directly to the `bulk` package was rejected because stdin reading is a transport concern, not a bulk-data-format concern.

---

### Decision 3: Interaction with existing flags when `--from-stdin` is set

**Decision**: Mutual exclusivity enforced at runtime, not at the flag level.

- **Group A**: If both `--file` and `--from-stdin` are provided, return an error: `"--file and --from-stdin are mutually exclusive"`.
- **Group B**: If `--from-stdin` is set, the individual field flags (`--display-name`, `--tracked`, etc.) are ignored — the JSON payload is the sole source of truth. The "at least one flag required" validation is skipped.
- **Group C**: `--from-stdin` sets an internal `skipConfirmation` bool, identical in effect to `--yes`.

**Rationale**: Cobra does not support native flag mutual exclusivity for this pattern. A runtime check is simple, explicit, and surfaces a clear error.

---

### Decision 4: Error handling for invalid or incomplete stdin JSON

**Decision**: Return structured, actionable errors:
- Syntax error: `--from-stdin: invalid JSON: <underlying error>`
- EOF/empty stdin: `--from-stdin: stdin is empty; pipe a JSON payload`
- Required fields missing (Group B): validate after unmarshal and report which fields are absent

**Rationale**: Matches the CLAUDE.md principle of "Graceful Degradation — actionable error messages."

---

### Decision 5: `consumers password set` with `--from-stdin`

**Decision**: When `--from-stdin` is set on `consumers password set`, the command reads a JSON object `{"password": "..."}` from stdin and sets the password without any interactive prompt. The `readPassword` function is bypassed entirely.

**Rationale**: The primary use case is agents/scripts setting passwords without terminal interaction. Reading a JSON payload is consistent with the rest of the feature's design.
