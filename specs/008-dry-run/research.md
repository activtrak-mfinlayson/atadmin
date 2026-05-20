# Research: Safe Exploration (--dry-run)

## Decision 1: Interception Point

**Decision**: Short-circuit inside `Client.doRequest()` in `internal/api/helpers.go`.

**Rationale**: `doRequest` has access to `method`, `path`, and `body` (the Go struct, before serialization). This is the single choke-point for all HTTP calls — adding the check here covers all ~50 mutating commands with zero per-command changes. The `body any` value is available before the `json.Marshal` call, so we can re-marshal it for the dry-run payload without re-reading a consumed stream.

**Alternatives considered**:
- *RoundTripper wrapper* — rejected because the body has already been serialized to a `bytes.Reader` by the time the transport sees it; re-reading it requires cloning the request body, adding complexity with no benefit
- *Per-command check in `internal/cmd/`* — rejected because it requires ~50 independent changes and risks omissions

---

## Decision 2: Client Fields for Dry-Run

**Decision**: Add `DryRun bool` and `Out io.Writer` to the `api.Client` struct.

**Rationale**: Mirrors the existing `verboseRoundTripper` pattern where `out io.Writer` is injected at construction time. Having `Out` on the client makes `doRequest` unit-testable without capturing `os.Stdout`.

**Alternatives considered**:
- *Global package variable* — rejected (global state violates CLAUDE.md guidance)
- *Context value* — rejected (implicit; harder to discover; context is for cancellation/deadlines)

---

## Decision 3: NewClient Signature

**Decision**: Add `dryRun bool` and `out io.Writer` parameters to `NewClient()`.

**Rationale**: All other client options (`verbose`, `verboseOut`) are already constructor parameters. Consistency is more important than minimizing parameter count.

---

## Decision 4: Flag Scope

**Decision**: Persistent root-level `--dry-run` flag, identical in scope to `--verbose`.

**Rationale**: The flag applies uniformly to all mutating subcommands. Declaring it once on the root with `PersistentFlags()` makes it available everywhere with no per-command boilerplate.

---

## Decision 5: GET Requests Under --dry-run

**Decision**: No interception for `GET` (and other safe methods).

**Rationale**: The spec explicitly targets "mutating commands". GET requests are idempotent and read-only; suppressing them would silently break commands that mix reads and writes (e.g., a `get` before a `set`).

---

## Decision 6: Synthetic Response

**Decision**: When dry-run intercepts, return `&http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))}`.

**Rationale**: Callers always call `checkResponse(resp)` after `doRequest`. A synthetic `200` passes that check without changes to caller code. The empty body satisfies the `defer resp.Body.Close()` pattern safely.

---

## Decision 7: `target` Field Value

**Decision**: Use the `path` string as-is from `doRequest`'s argument (e.g., `/admin/v1/clients/123`).

**Rationale**: The path is already path-only (no scheme/host), unambiguous, and matches what an agent would expect when cross-referencing the API docs. No transformation needed.

---

## Decision 8: `payload` for nil Body

**Decision**: Use JSON `null` when `body` is `nil`.

**Rationale**: DELETE requests with no request body have no payload. JSON `null` is the conventional representation of "no value" and is more explicit than omitting the field.

---

## Decision 9: Output Format

**Decision**: One JSON object per `doRequest` call, written as a single line to `Client.Out`.

**Rationale**: Bulk operations are a single HTTP request with a bulk payload — one line is correct. Single-line JSON is consistent with the `--json` output mode used throughout the CLI and is trivially parseable by `jq` or agents.
