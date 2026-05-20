# Research: Context Window Protections

## Decision: Where to implement field filtering

**Decision**: Client-side in the output layer (`internal/output`), not the API layer.  
**Rationale**: The remote API is external and not modifiable. Filtering after the full payload is received keeps the API client unchanged and the feature self-contained.  
**Alternatives considered**: API-level field projection (out of scope per spec); a new middleware layer (unnecessary; the output package already owns JSON serialization).

## Decision: Where to place FilterFields and SummaryResult

**Decision**: Add to `internal/output/output.go` as package-level functions alongside `Table`, `KeyValue`, and `JSON`.  
**Rationale**: The output package is already the single place all cmd files import for formatting. Adding there avoids a new package and keeps the dependency graph flat.  
**Alternatives considered**: A new `internal/filter` package (over-engineering for two functions); inlining per-command (code duplication across 11 files).

## Decision: Pagination safe-default value

**Decision**: 50 items.  
**Rationale**: Specified explicitly in the spec. Consistent with the existing `--page-size` default of 50 used in groups, clients, consumers, devices, and alarms commands.  
**Alternatives considered**: 25 (too conservative for most agent workflows), 100 (risks context overflow for large objects).

## Decision: total_items handling in --summary

**Decision**: Use `*int` (pointer) for `TotalItems`; omit the field (`omitempty`) when the API response does not include a total count.  
**Rationale**: Most endpoints return raw arrays without total counts. Forcing a total would require an extra API call or be misleading. Omitting it is honest and safe.  
**Alternatives considered**: Always include `total_items: -1` as a sentinel (confusing for agents), compute it from a second API call (doubles latency and complexity).

## Decision: --summary placement relative to --fields

**Decision**: `--summary` short-circuits before `--fields` is applied.  
**Rationale**: Summary mode returns aggregate metadata, not item data. Applying field filtering to a summary object has no meaning and would produce incorrect output.  
**Alternatives considered**: Let --summary + --fields coexist (semantically incoherent).

## Decision: MCP compatibility strategy

**Decision**: No changes to `internal/mcp/`. New flags are picked up automatically.  
**Rationale**: The mapper's `Walk` uses `cmd.Flags().VisitAll()` to enumerate all flags at startup. New flags registered on any command appear automatically in the next server start.  
**Alternatives considered**: Explicit MCP registration (unnecessary and violates the "zero-registration" invariant from feature 004).

## Decision: Scope — which commands get --summary

**Decision**: All list commands that return arrays (users, groups, clients, consumers, devices, alarms, auditlog, agents). Commands that return non-paginated flat arrays (signals, schedules, apikeys) get `--fields` but not `--summary` since `has_more` is always false and `total_items` is the full count returned (less useful).  
**Rationale**: Summary is most useful for paginated results. For flat arrays returned in a single call, the caller can simply `jq length` the result.  
**Alternatives considered**: Add `--summary` everywhere (low cost, low value for non-paginated results; deferred to later if requested).
