# Feature Specification: Identity API CLI Commands

**Feature Branch**: `003-identity-api-cli`  
**Created**: 2026-05-19  
**Status**: Draft  
**Input**: User description: "@docs/identity-swagger.json we'd like to add this api to the command line tool as well"

## Clarifications

### Session 2026-05-19

- Q: Should the top-level command group be `atadmin users`, `atadmin identities`, or `atadmin entities`? → A: `atadmin users` — operator-friendly noun, consistent with existing CLI conventions.
- Q: In non-TTY (script) mode, when `--revision` is omitted, should the CLI auto-fetch or require the flag explicitly? → A: Always auto-fetch silently in both TTY and non-TTY mode; `--revision` is an optional performance optimization only.
- Q: Should agent commands be top-level (`atadmin agents`) or nested under users (`atadmin users agents`)? → A: Top-level peer `atadmin agents`; scope is account-wide device inventory.
- Q: What is the default page size for `atadmin users list` when `--limit` is omitted? → A: No CLI-side default; omit the `limit` parameter and defer to the server's default.
- Q: How should the CLI handle 5xx errors and network timeouts? → A: Fail immediately with an actionable stderr message; no automatic retries on any command type.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List and Search Users (Priority: P1)

An administrator wants to browse and filter the identity entities (users) in their ActivTrak account from the command line. They need to see who is tracked, what groups they belong to, and their current activity status, without having to open the web UI.

**Why this priority**: Listing and searching users is the most fundamental operation against the Identity API. All other operations (edit, group management, bulk actions) begin with finding the target user. This delivers immediate value as a read-only audit/inspection tool and is the foundation every other story builds on.

**Independent Test**: Can be fully tested by running `atadmin users list` and `atadmin users list --filter tracked` against a live or mocked API, verifying that a table of identities is returned with correct columns and that filters reduce the result set.

**Acceptance Scenarios**:

1. **Given** valid credentials, **When** the operator runs `atadmin users list`, **Then** a table is printed showing each identity's ID, display name, primary group, status (active/inactive/unlicensed), and whether it is tracked.
2. **Given** a large account, **When** the operator runs `atadmin users list --filter tracked`, **Then** only tracked identities are returned.
3. **Given** a search term, **When** the operator runs `atadmin users list --search alice --search-type email`, **Then** only identities whose email matches "alice" are returned.
4. **Given** any `list` command, **When** `--json` flag is passed, **Then** the full `UsersResponse` JSON payload is written to stdout and no table is rendered.
5. **Given** a cursor from a previous page, **When** `--cursor <value>` is passed, **Then** the next page of results is returned.

---

### User Story 2 - Inspect a Single User (Priority: P1)

An administrator wants to see all details for a specific identity entity: their names, emails, UPNs, employee IDs, device logons, group memberships, associated agents, and the current revision number needed for safe updates.

**Why this priority**: Tied in priority with listing, because operators need to confirm identity details before taking any mutating action. The revision number exposed here is the prerequisite for all update and delete operations.

**Independent Test**: Can be tested independently by running `atadmin users get <id>` and verifying the key-value output includes `id`, `revision`, `displayName`, `emails`, `groups`, and `tracked` fields.

**Acceptance Scenarios**:

1. **Given** a valid entity ID, **When** the operator runs `atadmin users get 12345`, **Then** a key-value detail view is printed with all identity fields including the current revision.
2. **Given** an invalid or nonexistent ID, **When** the operator runs `atadmin users get 99999`, **Then** a clear error message is printed and the exit code is non-zero.
3. **Given** the `--json` flag, **When** the operator runs `atadmin users get 12345 --json`, **Then** the full `IdentityDetailsResponse` JSON is written to stdout.

---

### User Story 3 - Manage User Group Memberships (Priority: P2)

An administrator wants to add or remove an identity from one or more groups directly from the CLI. This is a common day-to-day admin task that currently requires navigating the web UI.

**Why this priority**: Group membership is the primary classification mechanism in ActivTrak. Automating group changes is high-value for bulk onboarding/offboarding scripts and complements the existing `groups clients add` functionality already in the tool.

**Independent Test**: Can be tested independently by adding a group to a user with `atadmin users groups add <userId> <groupId>` and then running `atadmin users get <userId>` to confirm the group appears, without touching any other user story functionality.

**Acceptance Scenarios**:

1. **Given** a valid user ID and group ID, **When** the operator runs `atadmin users groups add <userId> <groupId>`, **Then** the group is added to the identity and the updated identity detail is confirmed (no 409 conflict).
2. **Given** a stale revision, **When** the operator retries without refreshing first, **Then** a 409 Conflict error is surfaced with advice to re-fetch the user and retry.
3. **Given** a valid user ID and group ID already assigned, **When** the operator runs `atadmin users groups remove <userId> <groupId>`, **Then** the group is removed from the identity.
4. **Given** multiple group IDs, **When** `atadmin users groups add <userId> --group-ids 1,2,3` is run, **Then** all three groups are added in a single API call.

---

### User Story 4 - Update User Fields (Priority: P2)

An administrator wants to update scalar fields on an identity record — such as display name, timezone, or tracking state — without opening the web UI.

**Why this priority**: Editing individual user fields is a frequent admin task. Supporting at least display name, timezone, and tracked status enables meaningful automation. The optimistic concurrency model (revision-based) must be surfaced clearly so scripts do not silently clobber concurrent edits.

**Independent Test**: Can be tested independently by running `atadmin users update <id> --display-name "Alice Smith"` and verifying the returned entity shows the updated display name, without implementing any other story.

**Acceptance Scenarios**:

1. **Given** a valid user ID and a new display name, **When** the operator runs `atadmin users update <id> --display-name "Alice Smith"`, **Then** the field is patched and the updated entity is returned.
2. **Given** a valid user ID, **When** the operator runs `atadmin users update <id> --tracked=false`, **Then** the identity's tracked state is set to false.
3. **Given** a concurrent edit has happened since the operator last fetched the entity, **When** the operator submits an update, **Then** a 409 Conflict error is returned with a human-readable message directing the operator to re-fetch and retry.
4. **Given** the CLI is run without a `--revision` flag, **Then** the CLI automatically fetches the current revision before submitting the PATCH, so the operator does not need to manage revision numbers manually in interactive use.

---

### User Story 5 - Bulk Actions on Multiple Users (Priority: P3)

An administrator needs to apply the same action (start tracking, stop tracking, delete entity, delete data) across many identities at once, using a filtered list or explicit IDs.

**Why this priority**: Bulk operations are powerful but more complex to implement safely (revision fetching for each entity, partial-failure reporting). Deferring to P3 allows the single-entity patterns to stabilize first.

**Independent Test**: Can be tested independently by running `atadmin users bulk stop-tracking --ids 1,2,3` and checking the bulk action response for successes and failures.

**Acceptance Scenarios**:

1. **Given** a list of entity IDs, **When** the operator runs `atadmin users bulk start-tracking --ids 1,2,3`, **Then** all three entities are switched to tracked status and a summary of successes/failures is printed.
2. **Given** one entity in the list has a revision conflict, **When** the bulk action runs, **Then** the failure is reported per-entity without cancelling the rest of the batch.
3. **Given** `--json` is passed, **When** the bulk action completes, **Then** the `IdentityBulkActionResponse` JSON (with `successful` and `failures` arrays) is written to stdout.

---

### User Story 6 - List and Inspect Agent Entities (Priority: P3)

An administrator wants to list and inspect the agent (device) entities registered in the account — seeing their username, domain, alias, last activity, license status, and associated identity.

**Why this priority**: Agents represent the physical devices monitored by ActivTrak. Operators need to audit device inventory, identify orphaned agents, and understand which identity each device belongs to. Less urgent than user management because agents are rarely edited directly.

**Independent Test**: Can be tested independently by running `atadmin agents list` and verifying a table of agents with device-specific columns is returned.

**Acceptance Scenarios**:

1. **Given** valid credentials, **When** the operator runs `atadmin agents list`, **Then** a table is printed with agent username, domain, alias, last log date, and license status.
2. **Given** a `--filter usersWithAgents` flag, **When** the list command runs, **Then** only identities that have at least one associated agent are returned.
3. **Given** `--json`, **When** the command runs, **Then** the full `UsersResponse` JSON with nested agent detail is written to stdout.

---

### Edge Cases

- What happens when an entity's revision changes between the `get` and the `update` call in the same script? The CLI must surface the 409 with a clear message and non-zero exit code.
- What happens when `--search` returns zero results? An empty table with headers is printed; the exit code is 0.
- What happens when the Identity API base URL differs from the Admin API URL? The CLI must use a configurable base URL or derive it from the existing config.
- What happens when a bulk action partially fails? The CLI must print a mixed success/failure summary and exit with a non-zero code to signal partial failure.
- What happens when `--ids` and `--filter` are both supplied to a bulk command? One must take precedence; the CLI should document this and either reject the combination or prefer explicit IDs.
- What happens on a 500 Internal Server Error or network timeout? The CLI fails immediately with a non-zero exit code and an error on stderr that includes the status code; no retries are attempted.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST expose an `atadmin users` command group with subcommands: `list`, `get`, `update`, `delete`, `groups add`, `groups remove`, `bulk`.
- **FR-002**: The CLI MUST expose an `atadmin agents list` command that queries `GET /identity/v1/agents`.
- **FR-003**: `atadmin users list` MUST support `--filter <type>` (tracked, untracked, active30days, groupcount, nogroupcount, nodisplayname, orphans, unlicensed, activeUsers, inactiveUsers, pendingLicense), `--search <term>`, `--search-type <type>`, `--sort <field>`, `--sort-dir asc|desc`, `--limit <n>`, and `--cursor <token>` for pagination. When `--limit` is omitted, no limit parameter is sent to the server; the server's own default applies.
- **FR-004**: Every command that returns data MUST support `--json` to write raw API response JSON to stdout.
- **FR-005**: `atadmin users get <id>` MUST display all `IdentityDetailsResponse` fields in a human-readable key-value format, including the current `revision` number.
- **FR-006**: `atadmin users update <id>` MUST support flags `--display-name`, `--first-name`, `--last-name`, `--timezone`, `--tracked`. When `--revision` is not supplied, the CLI MUST automatically fetch the current revision before patching — this applies in both interactive (TTY) and script (non-TTY) mode. `--revision <n>` is an explicit override to skip the extra round-trip.
- **FR-007**: `atadmin users groups add <userId> <groupId>` and `atadmin users groups remove <userId> <groupId>` MUST handle revision fetching automatically in interactive mode.
- **FR-008**: `atadmin users bulk <action> --ids <id,...>` MUST support actions `start-tracking`, `stop-tracking`, `delete-data`, `delete-entity`. Each entity's revision MUST be fetched before the bulk request is sent.
- **FR-009**: All 409 Conflict responses MUST be surfaced as actionable error messages directing the operator to re-fetch the entity and retry.
- **FR-010**: `atadmin users delete <id>` MUST prompt for confirmation in interactive (TTY) mode and succeed silently in non-interactive mode when `--yes` is provided.
- **FR-011**: All error responses MUST include the HTTP status code and the API `message` field in the output to stderr; stdout remains clean for piping. On 5xx or network timeout, the CLI MUST fail immediately with a non-zero exit code and a human-readable error — no automatic retries on any command type.
- **FR-012**: The Identity API base path (`/identity/v1/...`) MUST be routed through the same configurable base URL used by the existing Admin API client, or a separate `ATADMIN_IDENTITY_BASE_URL` environment variable if the server differs.

### Key Entities

- **Identity (User)**: A logical person in the ActivTrak system. Key attributes: `id` (int64), `revision` (int64, required for mutations), `displayName`, `firstName`, `lastName`, `emails`, `upns`, `employeeIds`, `groups`, `deviceLogons`, `agents`, `tracked`, `status` (active/inactive/unlicensed), `timezone`, `created`, `updated`.
- **Agent**: A physical device/client associated with an identity. Key attributes: `userId`, `userName`, `logonDomain`, `alias`, `tracked`, `licenseStatus` (approved/pending/deleted), `lastLog`, `firstLog`.
- **Group** (reference): Already modelled in the codebase. Within the Identity context, a group attachment on an identity includes `groupId`, `groupName`, `groupType`.
- **Field**: A structured value with `value`, `id` (field-level identifier for targeted updates), and `source` (which system last set this value).
- **Revision**: An optimistic concurrency counter on each identity. All mutating operations (`PATCH`, `DELETE`, group add/remove) require passing the current revision and will fail with 409 if it has changed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator can list all identity entities in their account and pipe the JSON output to another tool without any non-JSON content on stdout.
- **SC-002**: An operator can inspect, update, and manage group memberships for a single user in under five commands, without opening the web UI.
- **SC-003**: All `users` and `agents` subcommands pass their `httptest`-based test suite with no API credentials required.
- **SC-004**: Revision conflict errors (409) result in a non-zero exit code and a stderr message that names the conflicting entity ID and advises the operator on how to retry.
- **SC-005**: Bulk operations on 100 entities complete and report a per-entity success/failure breakdown without leaving partial state unreported.
- **SC-006**: The new commands appear in `atadmin --help` and `atadmin users --help` output, fully described with flags and examples.

## Assumptions

- The Identity API is served from the same host as the Admin API, under the `/identity/v1/` path prefix; if this turns out to be a different host, an `ATADMIN_IDENTITY_BASE_URL` environment variable override will be introduced.
- The existing `api.Client` authentication (Bearer token via `http.RoundTripper`) applies unchanged to Identity API calls.
- Revision management is always automated: the CLI auto-fetches the current revision before any PATCH/DELETE regardless of TTY mode. `--revision <n>` is an optional flag to skip that round-trip when the caller has already fetched the revision.
- The `DELETE /identity/v1/entities/{id}` endpoint deletes the entity record; data deletion is a separate bulk action (`delete-data`). Both are scoped to this feature.
- The `/internal/v1/...` diagnostic endpoints (snapshots, changelogs, DNT diagnostics) are out of scope for this feature — they are internal service tooling, not operator-facing admin operations.
- Merge operations (`PATCH` with `merges` field, `/revision/{revision}/deleteClients`) are deferred to a follow-on feature due to their complexity and lower operator frequency.
- The `identifier/search` and `entity/search` (v1/v3) endpoints are not exposed in this feature; the `--search` and `--search-type` flags on `users list` cover the primary search use case.
- HRIS integration management (adding/removing emails, UPNs, employee IDs as individual `FieldDto` entries) is out of scope for v1; `users update` covers only the scalar fields (display name, first/last name, timezone, tracked).
