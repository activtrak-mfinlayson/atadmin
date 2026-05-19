# Feature Specification: ActivTrak Admin API CLI Wrapper

**Feature Branch**: `002-wrap-admin-api`
**Created**: 2026-05-18
**Status**: Draft
**Input**: User description: "wrap this api into our cli @docs/admin-swagger.json"

## Clarifications

### Session 2026-05-18

- Q: What file format(s) should the `--file` flag accept for bulk input operations? → A: Both JSON and CSV, auto-detected by file extension (`.json` → JSON, `.csv` → CSV)
- Q: How should the CLI respond to a rate-limit (429) response from the API? → A: Auto-retry with exponential backoff up to 3 times, then fail with an error showing retry count
- Q: Should the CLI support a flag for verbose/debug output showing request details? → A: Global `--verbose` flag printing HTTP method, URL, status code, and retry attempts to stderr
- Q: How should the CLI display single-object and nested-settings responses in the default (non-JSON) view? → A: Key-value pairs (`setting: value`), one per line, printed to stdout
- Q: Should the CLI support named configuration profiles for managing multiple accounts? → A: Named profiles via config — `atadmin --profile <name> <command>` selects a stored credential set; `atadmin auth login --profile <name>` adds a profile

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Client Management (Priority: P1)

An ActivTrak administrator needs to list, view, update, merge, delete, and restore tracked users (clients) from the command line to automate routine management tasks and integrate with scripting or DevOps workflows.

**Why this priority**: Clients are the most fundamental resource — they represent the end-users whose activity is tracked. Day-to-day management (alias changes, do-not-track overrides, user merges) is the most frequent operational task for admins.

**Independent Test**: Can be fully tested by running `atadmin clients list`, `atadmin clients get <id>`, `atadmin clients delete`, `atadmin clients merge`, `atadmin clients restore`, and `atadmin clients donottrack list/add/remove` — each producing correct, formatted output and making the expected API call.

**Acceptance Scenarios**:

1. **Given** valid stored credentials, **When** the user runs `atadmin clients list`, **Then** a formatted table is printed to stdout showing all clients with username, alias, and status; errors go to stderr
2. **Given** a client ID, **When** the user runs `atadmin clients get <id>`, **Then** the full client detail record is displayed
3. **Given** a `--json` flag, **When** any `clients` command is run, **Then** raw JSON from the API is written to stdout and no table formatting is applied
4. **Given** a 401 response, **When** any command is run, **Then** stderr shows: `Error: Unauthorized. Your token may have expired. Try running 'atadmin auth login'.`
5. **Given** a non-interactive (piped) environment, **When** `atadmin clients delete <id>` succeeds, **Then** only the deleted resource ID is printed to stdout

---

### User Story 2 - Group Management (Priority: P2)

An administrator needs to create, rename, delete, and manage membership of groups (collections of clients and devices) to control reporting boundaries and policy enforcement.

**Why this priority**: Groups are the second-most-common operational target — nearly all reporting and schedule features depend on correct group membership.

**Independent Test**: Can be fully tested by creating a group, adding members, listing members, removing members, and deleting the group — with each step verifiable independently.

**Acceptance Scenarios**:

1. **Given** valid credentials, **When** the user runs `atadmin groups list`, **Then** a table of groups with names and member counts is displayed
2. **Given** a group name, **When** the user runs `atadmin groups create <name>`, **Then** the group is created and its new ID is printed to stdout
3. **Given** a group ID and client ID, **When** the user runs `atadmin groups clients add --group <id> --client <id>`, **Then** the client is added to the group
4. **Given** a bulk import file, **When** the user runs `atadmin groups members import --file members.json`, **Then** all memberships in the file are applied

---

### User Story 3 - Consumer (Admin User) Management (Priority: P3)

An administrator needs to create, update, deactivate, and manage ActivTrak admin user accounts (consumers) including role assignments and password management.

**Why this priority**: Consumer management enables self-service onboarding and offboarding of admin users without requiring web UI access.

**Independent Test**: Can be fully tested by creating a consumer, updating their role, updating their password, and deleting them.

**Acceptance Scenarios**:

1. **Given** valid credentials, **When** the user runs `atadmin consumers list`, **Then** a table of all consumers with ID, username, and role is displayed
2. **Given** a consumer ID and new role, **When** the user runs `atadmin consumers role set --consumer <id> --role <role>`, **Then** the role is updated and success is confirmed
3. **Given** a consumer ID, **When** the user runs `atadmin consumers password set --consumer <id>`, **Then** the user is prompted to enter a new password (masked) in interactive mode

---

### User Story 4 - Account Settings Management (Priority: P4)

An administrator needs to view and update account-wide configuration settings (privacy, SSO, timezone, role access, schedule adherence, agent behavior) without logging into the ActivTrak web interface.

**Why this priority**: Settings changes are infrequent but high-impact; CLI access enables auditable, scriptable configuration management.

**Independent Test**: Can be fully tested by getting each setting type, updating a setting, and verifying the change is reflected in a subsequent get.

**Acceptance Scenarios**:

1. **Given** valid credentials, **When** the user runs `atadmin settings privacy get`, **Then** the current privacy settings are displayed
2. **Given** new privacy values, **When** the user runs `atadmin settings privacy set --activities <value>`, **Then** the privacy setting is updated
3. **Given** SSO eligibility is true, **When** the user runs `atadmin settings sso get`, **Then** SSO configuration is displayed with enabled/eligible status

---

### User Story 5 - Signals, Alarms & Schedules (Priority: P5)

An administrator needs to create, view, update, and delete activity alarms, notification signals, and work schedules to automate policy enforcement and reporting configuration.

**Why this priority**: These resources support policy automation; CLI access allows them to be version-controlled and deployed programmatically.

**Independent Test**: Can be fully tested by listing, creating, and deleting a signal, alarm, and schedule independently.

**Acceptance Scenarios**:

1. **Given** valid credentials, **When** the user runs `atadmin signals list`, **Then** all configured signals are displayed in a table
2. **Given** alarm parameters, **When** the user runs `atadmin alarms create`, **Then** the alarm is created and its ID is printed
3. **Given** a schedule ID, **When** the user runs `atadmin schedules users list --schedule <id>`, **Then** all users assigned to that schedule are displayed

---

### Edge Cases

- What happens when an API endpoint returns a paginated result with more pages? The CLI prompts the user in interactive mode or respects `--page` and `--page-size` flags in all modes.
- What happens when a bulk operation partially succeeds? Both the succeeded and failed items are reported, and the exit code is non-zero if any items failed.
- What happens when the configured API base URL is unreachable? The CLI prints a connection error to stderr with the attempted URL and exits non-zero.
- What happens when required flags are missing in non-interactive mode? The CLI fails immediately with a clear message listing the missing flags, never hanging for input.
- How are deprecated legacy endpoints handled? They are accessible but their help text is marked `[DEPRECATED]` and they warn on stderr when used.
- What happens when `--profile <name>` references a profile that does not exist in the config? The CLI prints an actionable error to stderr listing available profile names and exits non-zero.
- What happens when the API returns HTTP 429 Too Many Requests? The CLI automatically retries up to 3 times using exponential backoff; after exhausting retries it prints an error to stderr indicating the number of attempts made and exits non-zero.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide subcommands for all 11 resource groups: `accounts`, `alarms`, `apikeys`, `auditlog`, `clients`, `consumers`, `devices`, `groups`, `hrdc`, `schedules`, `signals` — structured as `atadmin <resource> <action>`
- **FR-002**: Every subcommand that retrieves data MUST support a `--json` flag (and `--format=json`) that outputs raw JSON to stdout, enabling pipe-friendly scripting
- **FR-003**: All data output (tables, JSON) MUST go to stdout; all diagnostic output (errors, warnings, progress) MUST go to stderr
- **FR-004**: All API error responses (400, 401, 403, 404, 5xx) MUST produce actionable error messages — not raw HTTP status codes — with guidance on resolution
- **FR-005**: The CLI MUST auto-detect interactive (TTY) vs. non-interactive (piped/CI) mode: use formatted tables and prompts in interactive mode; suppress formatting and fail fast on missing input in non-interactive mode
- **FR-006**: Paginated endpoints MUST expose `--page` (int) and `--page-size` (int) flags; the CLI MUST display total result counts when available
- **FR-007**: Mutation commands (create, update, delete) in non-interactive mode MUST output only the ID or key field of the affected resource on success — nothing else on stdout
- **FR-008**: All commands MUST support a configurable request timeout via a global `--timeout` flag (default: 30 seconds)
- **FR-009**: The CLI MUST automatically inject the stored bearer token into every API request without requiring manual header construction per command
- **FR-010**: Bulk operations (where the API provides a `/bulk` variant) MUST support accepting input from a file (`--file path`) in addition to inline flags; the file format is auto-detected by extension — `.json` for JSON, `.csv` for CSV; both formats MUST be supported
- **FR-011**: The `atadmin clients donottrack` command group MUST support listing, adding, removing, and updating do-not-track entries individually and in bulk
- **FR-012**: The `atadmin groups members` command group MUST support import and export of group membership data
- **FR-013**: The `atadmin settings` command group MUST cover all account settings sub-resources: privacy, sso, roleaccess, roledatefilter, timezone, agent activity duration, agent audit, passive time, schedule adherence, email auto-detection, identity settings, license approval mode
- **FR-014**: The `atadmin schedules` command group MUST distinguish reporting schedules from shift schedules, with separate sub-actions where the API differentiates them
- **FR-015**: Legacy endpoints (`/admin/legacy/`) MUST be accessible but their CLI commands MUST display a `[DEPRECATED]` notice in help text and a warning to stderr at runtime
- **FR-016**: The CLI MUST automatically retry requests that receive an HTTP 429 response, using exponential backoff, up to a maximum of 3 retry attempts; after exhausting retries it MUST print an error to stderr noting the retry count and exit non-zero
- **FR-017**: The CLI MUST support a global `--verbose` flag that, when set, writes the HTTP method, URL, response status code, and any retry attempts to stderr; verbose output MUST never appear on stdout
- **FR-018**: Commands that return a single object or settings structure (rather than a list) MUST display output as `key: value` pairs, one per line, in the default view; the `--json` flag overrides this and outputs raw JSON
- **FR-019**: The CLI MUST support named configuration profiles; `atadmin auth login --profile <name>` stores a named credential set, and `atadmin --profile <name> <command>` selects that credential set for the request; omitting `--profile` uses the default profile

### Key Entities

- **Client**: A tracked end-user; attributes include logon/username, alias, device associations, do-not-track status, and group memberships
- **Consumer**: An ActivTrak admin user; attributes include ID, username, role, SSO status, and viewable group scope
- **Device**: A monitored computer; attributes include device ID, hostname, agent status, and group memberships
- **Group**: A named collection of clients and/or devices used for reporting and policy scoping
- **Alarm**: A threshold-based alert with conditions, fields, and notification channels (email, webhook)
- **Signal**: A configurable notification definition (related to but distinct from Alarms)
- **Schedule**: A time-based schedule (reporting or shift type) with user assignments
- **ApiKey**: A credential for external integrations with the ActivTrak Public API
- **AuditLog**: An immutable record of administrative actions; supports date-range filtering and pagination

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 110 API endpoints defined in the Admin API OpenAPI specification are reachable via `atadmin` subcommands — zero endpoints excluded
- **SC-002**: Any command that returns list data can be piped to a downstream tool without manual parsing (verified by `atadmin clients list --json | jq '.[0].username'` succeeding)
- **SC-003**: In non-interactive mode, mutation commands produce no extra output on stdout beyond the resource ID or confirmation — enabling clean pipe chains such as `atadmin groups create "New Team" | atadmin groups clients add --group -`
- **SC-004**: All API error responses are translated into human-readable messages in under 10 words that include a suggested corrective action
- **SC-005**: The CLI never hangs waiting for stdin input in a non-TTY environment — confirmed by running any command with stdin closed (`< /dev/null`) and observing immediate exit with a clear error
- **SC-006**: A first-time user can discover and execute any command using only `--help` flags, without consulting external documentation

## Assumptions

- Authentication uses Bearer tokens stored per named profile by `atadmin auth login [--profile <name>]`; the default profile is used when `--profile` is omitted; the `ATADMIN_TOKEN` environment variable overrides the stored token for any profile
- The Admin API base URL defaults to `https://api.activtrak.com` and is overridable via `ATADMIN_API_URL` environment variable or `api_url` config file key
- All 110 endpoints are in scope for this feature; no endpoints are excluded
- The `/admin/legacy/notifications` endpoint is in scope but treated as deprecated in CLI help text
- The `/admin/v1/schedule/*` path group (singular) is considered a legacy alias of `/admin/v1/schedules/*` (plural); the CLI uses the plural form exclusively and maps to the correct path internally
- Paginated endpoints that do not specify a default will use a page size of 100 when the user does not provide `--page-size`
- The HRDC bulk import endpoint (`POST /hrdc/v1/bulk`) accepts either a JSON or CSV file as input, auto-detected by file extension
- The `atadmin clients get` command resolves by either numeric client ID or username, matching the API's dual-path support
- The `atadmin apikeys util backfill` commands are administrative utilities and will be included but placed under an `atadmin apikeys util` subgroup to indicate they are operational/maintenance commands, not everyday use
