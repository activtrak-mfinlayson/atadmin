# Tasks: ActivTrak Admin API CLI Wrapper

**Input**: Design documents from `specs/002-wrap-admin-api/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/cli-commands.md ✅

**Tests**: Included per CLAUDE.md guideline (test-first API clients using `httptest`; CLI commands via buffer tests).

**Organization**: Tasks grouped by user story to enable independent implementation and testing. Each phase is an independently deliverable increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel with other [P] tasks in the same phase
- **[Story]**: User story this task belongs to (US1–US5)

---

## Phase 1: Setup (New Infrastructure Packages)

**Purpose**: Create the shared packages that all resource commands depend on. Must complete before Phase 2.

- [ ] T001 Add `golang.org/x/term` dependency: run `go get golang.org/x/term` and verify `go.mod`/`go.sum` update
- [ ] T002 [P] Create `internal/tty/tty.go` with `IsTerminal() bool` wrapping `term.IsTerminal(int(os.Stdout.Fd()))`
- [ ] T003 [P] Create `internal/output/output.go` with `Table(out io.Writer, headers []string, rows [][]string)`, `KeyValue(out io.Writer, fields map[string]string)`, and `JSON(out io.Writer, v any) error` — `Table` uses `text/tabwriter`, `KeyValue` uses left-padded `fmt.Fprintf`
- [ ] T004 [P] Create `internal/output/output_test.go` testing Table, KeyValue, and JSON formatters against a `bytes.Buffer`
- [ ] T005 [P] Create `internal/bulk/bulk.go` with `ParseFile(path string) ([]map[string]any, error)` — auto-detects format by `filepath.Ext` (`.json` → `encoding/json`, `.csv` → `encoding/csv` with header row, other → error)
- [ ] T006 [P] Create `internal/bulk/bulk_test.go` with table-driven tests covering JSON file, CSV file, and unsupported extension error

---

## Phase 2: Foundational (Core Infrastructure Extensions)

**Purpose**: Extend the existing scaffold (config, API client, root command, models) to support profiles, retry, verbose, and all entity types. Blocks all user story work.

**⚠️ CRITICAL**: No user story phase can begin until this phase is complete.

- [ ] T007 Extend `internal/config/config.go`: add `LoadProfile(name string) (*Config, error)` that calls `v.Sub("profiles."+name)` — returns error listing available profiles if the sub-key is absent; add `ListProfiles() ([]string, error)` and `SaveProfile(name string, cfg *Config) error` (writes YAML with `0600` permissions); update existing `Load()` to call `LoadProfile("default")` for backward compatibility
- [ ] T008 [P] Update `internal/config/config_test.go`: add table-driven tests for `LoadProfile` (valid profile, missing profile, default fallback) and `SaveProfile` (file created, permissions `0600`)
- [ ] T009 Extend `internal/api/models.go`: add Go structs for all 9 data-model entities (Client, DNTEntry, Consumer, Device, Group, GroupMember, Alarm, Condition, Channel, Signal, Schedule, ApiKey, AuditLog) and all account settings sub-types from `data-model.md` — use JSON struct tags matching the Admin API field names from `docs/admin-swagger.json`
- [ ] T010 Extend `internal/api/client.go`: add `retryRoundTripper` struct (wraps inner transport, retries on HTTP 429 up to `maxRetry=3` times with exponential backoff `2^attempt` seconds using `time.Sleep`); add `verboseRoundTripper` struct (logs `> METHOD URL` and `< STATUS` to a configurable `io.Writer` stderr when enabled); update `NewClient(baseURL, token, version string, verbose bool) (*Client, error)` to chain `verboseRoundTripper → retryRoundTripper → authRoundTripper → http.DefaultTransport`; add `Timeout time.Duration` field to Client and apply it to the http.Client
- [ ] T011 [P] Update `internal/api/client_test.go`: add tests for `retryRoundTripper` (verifies 3 retries on 429, success on non-429, backoff timing order) and `verboseRoundTripper` (verifies log lines written to buffer when verbose=true, silent when false)
- [ ] T012 Extend `internal/cmd/root.go`: add `--profile` persistent flag (string, default `"default"`, env `ATADMIN_PROFILE`); add `--verbose` persistent flag (bool); add `PersistentPreRunE` that resolves config via `config.LoadProfile(profileFlag)` and constructs the API client, storing both in cobra's context or a package-level variable accessible to all subcommands; register stubs for all 11 resource command groups by calling `root.AddCommand(...)` — the stubs need not be implemented yet but must compile

---

## Phase 3: User Story 1 — Client Management (Priority: P1) 🎯 MVP

**Goal**: Full CRUD for tracked users including aliases, merges, restores, and Do Not Track management.

**Independent Test**: `atadmin clients list`, `atadmin clients get <id>`, `atadmin clients delete`, `atadmin clients merge`, `atadmin clients donottrack list` each return correct output and make the right API call.

### Implementation

- [ ] T013 [US1] Create `internal/api/clients.go`: implement `ListClients(ctx, page, pageSize int) ([]Client, error)`, `GetClientByID(ctx, id int) (*Client, error)`, `GetClientByUsername(ctx, username string) (*Client, error)`, `UpdateClient(ctx, id int, alias string) error`, `DeleteClients(ctx, ids []int) error`, `RestoreClients(ctx, ids []int) error`
- [ ] T014 [US1] Add to `internal/api/clients.go`: `MergeUsers(ctx, sourceID, targetID int) error`, `MergeUsersBulk(ctx, records []map[string]any) error`, `UnmergeUsersBulk(ctx, records []map[string]any) error`
- [ ] T015 [US1] Add to `internal/api/clients.go`: `UpdateAlias(ctx, id int, alias string) error`, `UpdateAliasBulk(ctx, records []map[string]any) error`, `ClientHealth(ctx) (int, error)`
- [ ] T016 [US1] Add to `internal/api/clients.go`: `ListDoNotTrack(ctx) ([]DNTEntry, error)`, `AddDoNotTrack(ctx, domain, username string) error`, `RemoveDoNotTrack(ctx, ids []int) error`, `UpdateDoNotTrack(ctx, id int, domain, username string) error`, `AddDoNotTrackBulk(ctx, records []map[string]any) error`, `RemoveDoNotTrackBulk(ctx, ids []int) error`, `MarkGlobalUser(ctx, ids []int) error`
- [ ] T017 [US1] Create `internal/api/clients_test.go`: table-driven `httptest.NewServer` tests for every method — verify HTTP method, path, request body, and that the response is correctly deserialized
- [ ] T018 [US1] Create `internal/cmd/clients.go`: implement `clients list` (table: username, alias, status; flags: `--page`, `--page-size`, `--json`), `clients get <id|username>` (key-value output, auto-detect int vs string arg), `clients health` (prints active count)
- [ ] T019 [US1] Add to `internal/cmd/clients.go`: `clients update <id> --alias <str>`, `clients delete --ids <id,...>`, `clients restore --ids <id,...>`, `clients merge --source <id> --target <id>`, `clients merge bulk --file <path>`, `clients unmerge bulk --file <path>`
- [ ] T020 [US1] Add to `internal/cmd/clients.go`: `clients alias set --id <id> --alias <str>`, `clients alias bulk --file <path>`
- [ ] T021 [US1] Add to `internal/cmd/clients.go`: `clients donottrack list`, `clients donottrack add --domain <str> --username <str>`, `clients donottrack remove --ids <id,...>`, `clients donottrack update --id <id> --domain <str> --username <str>`, `clients donottrack add-bulk --file <path>`, `clients donottrack remove-bulk --file <path>`, `clients donottrack global-user --ids <id,...>`
- [ ] T022 [US1] Create `internal/cmd/clients_test.go`: buffer tests for `clients list` (table and JSON output), `clients get`, `clients delete` (stdout ID only in non-TTY), `clients donottrack list`; mock API client via interface or `httptest` local server
- [ ] T023 [US1] Wire clients command: add `newClientsCmd()` to root in `internal/cmd/root.go`

**Checkpoint**: `go test ./internal/api/... ./internal/cmd/...` passes; `atadmin clients list --help` works.

---

## Phase 4: User Story 2 — Group Management (Priority: P2)

**Goal**: Full CRUD for groups and their memberships, including bulk import/export.

**Independent Test**: `atadmin groups list`, `atadmin groups create <name>`, `atadmin groups members import --file members.json`, and `atadmin groups clients add` each succeed independently.

### Implementation

- [ ] T024 [US2] Create `internal/api/groups.go`: `ListGroups(ctx, page, pageSize int) ([]Group, error)`, `GetGroupSummary(ctx) ([]Group, error)`, `GetGroup(ctx, id int) (*Group, error)`, `SearchGroups(ctx, prefix string) ([]Group, error)`, `CreateGroup(ctx, name string) (int, error)`, `RenameGroup(ctx, id int, name string) error`, `DeleteGroups(ctx, ids []int) error`
- [ ] T025 [US2] Add to `internal/api/groups.go`: `ListMembers(ctx, page, pageSize int) ([]GroupMember, error)`, `ListGroupMembers(ctx, groupID int) ([]GroupMember, error)`, `AddMembers(ctx, groupID, memberID int, memberType string) error`, `RemoveMembers(ctx, groupID, memberID int) error`, `ExportMembers(ctx) ([]byte, error)`, `ImportMembers(ctx, records []map[string]any) error`, `GetMembership(ctx, groupID, memberID int, memberType string) (*GroupMember, error)`
- [ ] T026 [US2] Add to `internal/api/groups.go`: `AddClientsToGroup(ctx, groupID int, clientIDs []int) error`, `RemoveClientsFromGroup(ctx, groupID int, clientIDs []int) error`, `AddDevicesToGroup(ctx, groupID int, deviceIDs []int) error`, `RemoveDevicesFromGroup(ctx, groupID int, deviceIDs []int) error`
- [ ] T027 [US2] Create `internal/api/groups_test.go`: `httptest` table-driven tests for all groups API methods
- [ ] T028 [US2] Create `internal/cmd/groups.go`: `groups list` (table: id, name, member count), `groups summary`, `groups get <id>` (key-value), `groups search <prefix>`, `groups create <name>` (prints new ID), `groups rename <id> --name <str>`, `groups delete --ids <id,...>`
- [ ] T029 [US2] Add to `internal/cmd/groups.go`: `groups members list` (with optional `--group <id>`), `groups members add --group <id> --member <id> --type client|device`, `groups members remove --group <id> --member <id>`, `groups members export --output <path>`, `groups members import --file <path>`, `groups membership get <group-id> <type> <member-id>`
- [ ] T030 [US2] Add to `internal/cmd/groups.go`: `groups clients add <group-id> --ids <id,...>`, `groups clients remove <group-id> --ids <id,...>`, `groups devices add <group-id> --ids <id,...>`, `groups devices remove <group-id> --ids <id,...>`
- [ ] T031 [US2] Create `internal/cmd/groups_test.go`: buffer tests for `groups list`, `groups create`, `groups members import --file`, error on missing profile
- [ ] T032 [US2] Wire groups command: add `newGroupsCmd()` to root in `internal/cmd/root.go`

**Checkpoint**: `atadmin groups list` and `atadmin groups members import --file members.json` succeed independently.

---

## Phase 5: User Story 3 — Consumer Management & Auth (Priority: P3)

**Goal**: Full CRUD for admin users including role/password management, plus completed `auth login` with profile support.

**Independent Test**: `atadmin consumers list`, `atadmin consumers role set`, `atadmin consumers password set` (interactive), and `atadmin auth login --profile staging` each succeed independently.

### Implementation

- [ ] T033 [US3] Create `internal/api/consumers.go`: `ListConsumers(ctx, page, pageSize int) ([]Consumer, error)`, `GetConsumer(ctx, id int) (*Consumer, error)`, `CreateConsumers(ctx, records []map[string]any) error`, `PatchConsumers(ctx, records []map[string]any) error`, `DeleteConsumers(ctx, ids []int) error`, `DeleteConsumersBulk(ctx, ids []int) error`
- [ ] T034 [US3] Add to `internal/api/consumers.go`: `SetConsumerRole(ctx, consumerID int, role string) error`, `SetConsumerPassword(ctx, consumerID int, password string) error`, `SetConsumerSSO(ctx, consumerID int, useSSO bool) error`, `AddViewableGroups(ctx, consumerID int, groupIDs []int) error`, `CreateChromeUsersBulk(ctx, records []map[string]any) error`
- [ ] T035 [US3] Create `internal/api/consumers_test.go`: `httptest` table-driven tests for all consumer methods
- [ ] T036 [US3] Create `internal/cmd/consumers.go`: `consumers list` (table: id, username, role), `consumers get <id>` (key-value), `consumers create --file <path>`, `consumers update --file <path>`, `consumers delete --ids <id,...>`, `consumers delete bulk --file <path>`
- [ ] T037 [US3] Add to `internal/cmd/consumers.go`: `consumers role set <id> --role <str>`, `consumers password set <id>` (prompts masked input in TTY; fails fast with error if non-TTY), `consumers sso set --consumer <id> --use-sso <bool>`, `consumers groups add <id> --group-ids <id,...>`, `consumers chrome-users import --file <path>`
- [ ] T038 [US3] Create `internal/cmd/consumers_test.go`: buffer tests for `consumers list`, `consumers role set`, `consumers delete` (ID-only stdout in non-TTY)
- [ ] T039 [US3] Implement `auth login` in `internal/cmd/auth.go`: add `--profile` flag; open browser to token generation URL using `os/exec`; prompt for token with masked echo using `golang.org/x/term`; validate token by calling `GET /admin/v1/accounts/ping`; save profile via `config.SaveProfile(name, cfg)` with `0600` permissions; print success message to stderr
- [ ] T040 [US3] Create/update `internal/cmd/auth_test.go`: test `auth login` with a mock ping server — verify profile saved, error on bad token, error on missing TTY for password prompt
- [ ] T041 [US3] Wire consumers command: add `newConsumersCmd()` to root in `internal/cmd/root.go`

**Checkpoint**: `atadmin consumers list` works; `atadmin auth login --profile staging` saves credentials and validates via ping.

---

## Phase 6: User Story 4 — Account Settings (Priority: P4)

**Goal**: View and update all 21 account-level settings sub-resources via `atadmin settings <sub-resource> get|set`.

**Independent Test**: `atadmin settings privacy get`, `atadmin settings sso get`, `atadmin settings role-access set` (from file) each succeed independently.

### Implementation

- [ ] T042 [US4] Create `internal/api/accounts.go`: `GetPrivacy(ctx) (map[string]any, error)`, `UpdatePrivacy(ctx, body map[string]any) error`, `GetSSO(ctx) (map[string]any, error)`, `UpdateSSO(ctx, body map[string]any) error`, `GetSSOEnabled(ctx) (bool, error)`, `GetSSOEligible(ctx) (bool, error)`, `Ping(ctx) error`
- [ ] T043 [US4] Add to `internal/api/accounts.go`: `GetRoleAccess(ctx) ([]map[string]any, error)`, `SetRoleAccess(ctx, body []map[string]any) error`, `ResetRoleAccess(ctx) error`, `GetRoleDateFilter(ctx) ([]map[string]any, error)`, `SetRoleDateFilter(ctx, body []map[string]any) error`
- [ ] T044 [US4] Add to `internal/api/accounts.go`: `GetTimezone(ctx) (map[string]any, error)`, `UpdateTimezone(ctx, body map[string]any) error`, `ListTimezones(ctx) ([]map[string]any, error)`, `GetLocalTimezone(ctx) (map[string]any, error)`, `UpdateLocalTimezone(ctx, body map[string]any) error`
- [ ] T045 [US4] Add to `internal/api/accounts.go`: agent duration methods (GetAgentDuration, AddAgentDuration, UpdateAgentDuration, DeleteAgentDuration — both v1 and v2 variants per swagger), `GetAgentAudit(ctx) (map[string]any, error)`, `UpdateAgentAudit(ctx, body map[string]any) error`
- [ ] T046 [US4] Add to `internal/api/accounts.go`: `GetPassiveTime(ctx) (map[string]any, error)`, `UpdatePassiveTime(ctx, body map[string]any) error`, `BulkUpdatePassiveTime(ctx, records []map[string]any) error`, `GetScheduleAdherence(ctx) (map[string]any, error)`, `UpdateScheduleAdherence(ctx, body map[string]any) error`
- [ ] T047 [US4] Add to `internal/api/accounts.go`: `GetEmailAutoDetect(ctx) (map[string]any, error)`, `UpdateEmailAutoDetect(ctx, body map[string]any) error`, `GetIdentityMatch(ctx) (map[string]any, error)`, `UpdateIdentityMatch(ctx, body map[string]any) error`, `GetIdentityThreshold(ctx) (map[string]any, error)`, `UpdateIdentityThreshold(ctx, body map[string]any) error`
- [ ] T048 [US4] Add to `internal/api/accounts.go`: `GetLicenseApproval(ctx) (map[string]any, error)`, `UpdateLicenseApproval(ctx, body map[string]any) error`, `GetMSPOverage(ctx) (map[string]any, error)`, `UpdateMSPOverage(ctx, body map[string]any) error`, `DeleteMSPOverage(ctx) error`, `GetHRIS(ctx) (map[string]any, error)`, `GetAcademyURL(ctx) (string, error)`, `GetAcademyWorkRampURL(ctx) (string, error)`
- [ ] T049 [US4] Create `internal/api/accounts_test.go`: `httptest` tests covering all accounts API methods
- [ ] T050 [US4] Create `internal/cmd/accounts.go`: top-level `settings` command; `settings privacy get` (key-value), `settings privacy set --file <path>`, `settings sso get`, `settings sso set --file <path>`, `settings sso enabled`, `settings sso eligible`, `settings ping`
- [ ] T051 [US4] Add to `internal/cmd/accounts.go`: `settings role-access get` (table: resource, roles), `settings role-access set --file <path>`, `settings role-access reset`, `settings role-date-filter get`, `settings role-date-filter set --file <path>`
- [ ] T052 [US4] Add to `internal/cmd/accounts.go`: `settings timezone get`, `settings timezone set --file <path>`, `settings timezones list`, `settings local-timezone get`, `settings local-timezone set --file <path>`
- [ ] T053 [US4] Add to `internal/cmd/accounts.go`: agent duration, agent audit, passive-time, schedule-adherence, email-autodetect, identity-match, identity-threshold, license-approval, msp-overage, hris get, academy url, academy workramp-url sub-commands (each with get/set where applicable, matching the contract in `contracts/cli-commands.md`)
- [ ] T054 [US4] Create `internal/cmd/accounts_test.go`: buffer tests for `settings privacy get` (key-value output), `settings role-access get` (table), `settings ping` (success/failure message to stderr)
- [ ] T055 [US4] Wire settings command: add `newSettingsCmd()` to root in `internal/cmd/root.go`

**Checkpoint**: `atadmin settings privacy get` and `atadmin settings sso get` return key-value output; `atadmin settings role-access set --file data.json` updates settings.

---

## Phase 7: User Story 5 — Signals, Alarms & Schedules (Priority: P5)

**Goal**: Create, view, update, and delete activity alarms, notification signals, and work schedules.

**Independent Test**: `atadmin signals list`, `atadmin alarms create`, and `atadmin schedules users list --schedule <id>` each succeed independently.

### Signals

- [ ] T056 [P] [US5] Create `internal/api/signals.go`: `ListSignals(ctx) ([]Signal, error)`, `GetNotifications(ctx) ([]Signal, error)`, `CreateSignal(ctx, body map[string]any) (int, error)`, `UpdateSignal(ctx, body map[string]any) error`, `DeleteSignal(ctx, id int) error`
- [ ] T057 [P] [US5] Create `internal/api/signals_test.go`: `httptest` tests for all signal methods
- [ ] T058 [P] [US5] Create `internal/cmd/signals.go`: `signals list` (table: id, name, type, enabled), `signals create --file <path>`, `signals update --file <path>`, `signals delete <id>` (prints ID on success in non-TTY)
- [ ] T059 [P] [US5] Create `internal/cmd/signals_test.go`: buffer tests for list (table/JSON) and delete (stdout ID)
- [ ] T060 [US5] Wire signals command: add `newSignalsCmd()` to root in `internal/cmd/root.go`

### Alarms

- [ ] T061 [P] [US5] Create `internal/api/alarms.go`: `ListAlarms(ctx, page, pageSize int) ([]Alarm, error)`, `GetAlarm(ctx, id int) (*Alarm, error)`, `GetAlarmDetails(ctx, id int) (map[string]any, error)`, `SaveAlarms(ctx, body map[string]any) error`, `SaveAlarm(ctx, body map[string]any) error`, `DeleteAlarm(ctx, id int) error`, `GetAlarmConditions(ctx) ([]Condition, error)`, `GetAlarmFields(ctx) ([]Field, error)`
- [ ] T062 [P] [US5] Create `internal/api/alarms_test.go`: `httptest` tests for all alarm methods
- [ ] T063 [P] [US5] Create `internal/cmd/alarms.go`: `alarms list` (table: id, name, type, enabled), `alarms get <id>` (key-value), `alarms details <id>`, `alarms create --file <path>`, `alarms update --file <path>`, `alarms delete <id>`, `alarms conditions`, `alarms fields`
- [ ] T064 [P] [US5] Create `internal/cmd/alarms_test.go`: buffer tests for list (table/JSON) and conditions output
- [ ] T065 [US5] Wire alarms command: add `newAlarmsCmd()` to root in `internal/cmd/root.go`

### Schedules

- [ ] T066 [P] [US5] Create `internal/api/schedules.go`: `ListSchedules(ctx) ([]Schedule, error)`, `GetSchedule(ctx, id int) (*Schedule, error)`, `CreateSchedule(ctx, body map[string]any) (int, error)`, `DeleteSchedule(ctx, id int) error`
- [ ] T067 [P] [US5] Add to `internal/api/schedules.go`: `GetReportingDefault(ctx) (*Schedule, error)`, `SetReportingDefault(ctx, scheduleID int) error`, `GetShiftDefault(ctx) (*Schedule, error)`, `SetShiftDefault(ctx, scheduleID int) error`; reporting users: `GetReportingUsers(ctx) ([]map[string]any, error)`, `RemoveReportingUsers(ctx, ids []int) error`; shift users: `GetShiftUsers(ctx) ([]map[string]any, error)`, `RemoveShiftUsers(ctx, ids []int) error`
- [ ] T068 [P] [US5] Add to `internal/api/schedules.go`: `GetScheduleUsers(ctx, scheduleID int) ([]map[string]any, error)`, `SetScheduleUsers(ctx, scheduleID int, userIDs []int) error`, `MoveUserToSchedule(ctx, scheduleID, userID int) error`, `GetUserReportingSchedule(ctx, userID int) (*Schedule, error)`, `GetUserShiftSchedule(ctx, userID int) (*Schedule, error)`, `RemoveUserFromReportingSchedules(ctx, userID int) error`, `RemoveUserFromShiftSchedules(ctx, userID int) error`
- [ ] T069 [P] [US5] Create `internal/api/schedules_test.go`: `httptest` tests for all schedule methods
- [ ] T070 [P] [US5] Create `internal/cmd/schedules.go`: `schedules list`, `schedules get <id>`, `schedules create --file <path>`, `schedules delete <id>`, `schedules reporting default get/set`, `schedules reporting users list/remove`, `schedules shift default get/set`, `schedules shift users list/remove`
- [ ] T071 [P] [US5] Add to `internal/cmd/schedules.go`: `schedules users list <id>`, `schedules users set <id> --ids <id,...>`, `schedules user move --schedule <id> --user <id>`, `schedules user get <user-id> reporting|shift`, `schedules user remove <user-id> reporting|shift`
- [ ] T072 [P] [US5] Create `internal/cmd/schedules_test.go`: buffer tests for list (table/JSON), schedules user get, default schedule get
- [ ] T073 [US5] Wire alarms and schedules commands: add `newSchedulesCmd()` to root in `internal/cmd/root.go`; confirm `newSignalsCmd()`, `newAlarmsCmd()`, `newSchedulesCmd()` all registered

**Checkpoint**: `atadmin signals list`, `atadmin alarms create --file alarm.json`, and `atadmin schedules users list --schedule 1` each work independently.

---

## Phase 8: FR-001 Coverage — Remaining Resource Groups

**Purpose**: Complete the remaining 4 resource groups (Devices, ApiKeys, AuditLog, HRDC) and the deprecated Notifications alias to reach 100% endpoint coverage per SC-001.

### Devices

- [ ] T074 [P] Create `internal/api/devices.go`: `ListDevices(ctx, page, pageSize int) ([]Device, error)`, `GetDevice(ctx, id int) (*Device, error)`, `DeleteDevices(ctx, ids []int) error`, `RestoreDevices(ctx, ids []int) error`, `UninstallDevice(ctx, ids []int) error`
- [ ] T075 [P] Create `internal/api/devices_test.go`: `httptest` tests for all device methods
- [ ] T076 [P] Create `internal/cmd/devices.go`: `devices list` (table: id, hostname, agent status), `devices get <id>` (key-value), `devices delete --ids <id,...>`, `devices restore --ids <id,...>`, `devices uninstall --ids <id,...>`
- [ ] T077 [P] Create `internal/cmd/devices_test.go`: buffer tests for list and delete
- [ ] T078 Wire devices command: add `newDevicesCmd()` to root in `internal/cmd/root.go`

### API Keys

- [ ] T079 [P] Create `internal/api/apikeys.go`: `ListAPIKeys(ctx) ([]ApiKey, error)`, `CreateAPIKey(ctx, name string) (*ApiKey, error)`, `UpdateAPIKey(ctx, id int, name string) error`, `DeleteAPIKey(ctx, id int) error`, `BackfillAllAPIKeys(ctx) error`, `BackfillAPIKey(ctx, id int) error`
- [ ] T080 [P] Create `internal/api/apikeys_test.go`: `httptest` tests for all API key methods
- [ ] T081 [P] Create `internal/cmd/apikeys.go`: `apikeys list` (table: id, name, key prefix, last used), `apikeys create --name <str>`, `apikeys update --id <id> --name <str>`, `apikeys delete <id>`, `apikeys util backfill all`, `apikeys util backfill <id>`
- [ ] T082 [P] Create `internal/cmd/apikeys_test.go`: buffer tests for list and create (prints key ID on stdout in non-TTY)
- [ ] T083 Wire apikeys command: add `newAPIKeysCmd()` to root in `internal/cmd/root.go`

### Audit Log

- [ ] T084 [P] Create `internal/api/auditlog.go`: `ListAuditLogs(ctx context.Context, from, to string, filters, sortCol string, sortDesc bool, page, pageSize int) ([]AuditLog, error)`, `GetAttachment(ctx context.Context, attachmentID string) ([]byte, error)`
- [ ] T085 [P] Create `internal/api/auditlog_test.go`: `httptest` tests for ListAuditLogs (verify query params correctly set) and GetAttachment
- [ ] T086 [P] Create `internal/cmd/auditlog.go`: `auditlog list` (table: id, action, actor, timestamp; flags: `--from`, `--to`, `--filters`, `--sort`, `--desc`, `--page`, `--page-size`), `auditlog attachment get <id>` (writes bytes to stdout or `--output <path>`)
- [ ] T087 [P] Create `internal/cmd/auditlog_test.go`: buffer tests for list with date filters, attachment get
- [ ] T088 Wire auditlog command: add `newAuditLogCmd()` to root in `internal/cmd/root.go`

### HRDC

- [ ] T089 [P] Create `internal/api/hrdc.go`: `HRDCPing(ctx) error`, `HRDCBulkImport(ctx context.Context, records []map[string]any) error`
- [ ] T090 [P] Create `internal/api/hrdc_test.go`: `httptest` tests for ping and bulk import (verify JSON body sent)
- [ ] T091 [P] Create `internal/cmd/hrdc.go`: `hrdc ping` (prints latency to stderr), `hrdc import --file <path>` (accepts `.json` or `.csv`, uses `bulk.ParseFile`; prints record count on success)
- [ ] T092 [P] Create `internal/cmd/hrdc_test.go`: buffer tests for ping and import with both JSON and CSV files
- [ ] T093 Wire hrdc command: add `newHRDCCmd()` to root in `internal/cmd/root.go`

### Deprecated Notifications

- [ ] T094 Create `internal/cmd/notifications.go`: `notifications list [DEPRECATED]` — calls `GET /admin/legacy/notifications`, prints `[DEPRECATED] Use 'atadmin signals list' instead` to stderr, then displays results in same table format as signals list
- [ ] T095 Wire notifications command: add `newNotificationsCmd()` to root in `internal/cmd/root.go`; mark command deprecated via `cobra.Command.Deprecated` field

**Checkpoint**: `go build -o bin/atadmin ./cmd/atadmin` succeeds; `atadmin --help` shows all 11 resource groups plus `auth` and `notifications`.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Validate full coverage, fix issues, and ensure all quality bars are met.

- [ ] T096 Run `go test ./...` — fix any failing tests before proceeding
- [ ] T097 Run `go build -o bin/atadmin ./cmd/atadmin` — fix any compilation errors
- [ ] T098 Run `golangci-lint run` — fix all lint issues
- [ ] T099 Verify SC-001 coverage: write a script (or manually verify) that cross-references all paths in `docs/admin-swagger.json` against `contracts/cli-commands.md` to confirm all 110 endpoints are mapped; document any gaps
- [ ] T100 [P] Validate error message contract: for each HTTP status code (400, 401, 403, 404, 429, 5xx, timeout, connection error), write or verify a test that the message matches the contract in `contracts/cli-commands.md`
- [ ] T101 [P] Validate TTY behavior: write an integration test (or manual verification) that runs `atadmin clients list < /dev/null` and confirms immediate non-zero exit, not a hang
- [ ] T102 Update `README.md` with usage examples covering auth login, clients, groups, consumers, settings, and signals commands

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately; all tasks parallelizable
- **Phase 2 (Foundational)**: Depends on Phase 1 completion — **BLOCKS all user story phases**
- **Phase 3–7 (User Stories)**: All depend on Phase 2; can proceed in priority order P1→P5 or in parallel if multiple developers
- **Phase 8 (Coverage)**: Can begin after Phase 2; entirely independent of Phases 3–7; all tasks parallelizable
- **Phase 9 (Polish)**: Depends on all prior phases complete

### User Story Dependencies

- **US1 (P1)**: No dependency on other stories — start after Phase 2
- **US2 (P2)**: No dependency on US1 — start after Phase 2
- **US3 (P3)**: No dependency on US1/US2 — start after Phase 2
- **US4 (P4)**: No dependency on US1/US2/US3 — start after Phase 2
- **US5 (P5)**: All three sub-streams (signals, alarms, schedules) are internally parallel

### Within Each User Story

- API layer (`internal/api/<resource>.go`) → CLI layer (`internal/cmd/<resource>.go`) → Wire to root
- API tests can be written alongside the API implementation
- CLI tests written alongside CLI implementation

---

## Parallel Execution Examples

```bash
# Phase 1 — all tasks in parallel:
Task: "Create internal/tty/tty.go"
Task: "Create internal/output/output.go"
Task: "Create internal/bulk/bulk.go"

# Phase 2 — partial parallelism:
Task: "Extend config.go with profile support"         # sequential (unblocks T008)
Task: "Extend api/models.go with all entity structs"  # [P] alongside config
Task: "Extend api/client.go with RoundTripper chain"  # sequential (needs T009)
Task: "Update config_test.go"                         # [P] after T007

# Phase 7 — US5 signals/alarms/schedules in parallel:
Task: "Create api/signals.go + test + cmd/signals.go"
Task: "Create api/alarms.go + test + cmd/alarms.go"
Task: "Create api/schedules.go + test + cmd/schedules.go"

# Phase 8 — all resource groups in parallel:
Task: "Devices API + cmd"
Task: "ApiKeys API + cmd"
Task: "AuditLog API + cmd"
Task: "HRDC API + cmd"
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1: Setup packages
2. Complete Phase 2: Foundational (CRITICAL — blocks all)
3. Complete Phase 3: US1 Client Management
4. **STOP and VALIDATE**: `atadmin clients list`, `atadmin clients get`, `atadmin clients donottrack list` work end-to-end
5. Demo and deploy MVP

### Incremental Delivery

1. Phase 1 + Phase 2 → Infrastructure ready
2. Phase 3 (Clients) → MVP: most common operational task delivered
3. Phase 4 (Groups) → Add group management
4. Phase 5 (Consumers + Auth) → Add admin user management + real auth
5. Phase 6 (Settings) → Add account configuration management
6. Phase 7 (Signals/Alarms/Schedules) → Add alerting and scheduling
7. Phase 8 (Remaining resources) → 100% API coverage
8. Phase 9 (Polish) → Production-ready

### Parallel Team Strategy

With multiple developers available after Phase 2 completes:
- Developer A: Phase 3 (Clients — P1)
- Developer B: Phase 4 (Groups — P2)
- Developer C: Phase 8 (Devices, ApiKeys, AuditLog, HRDC — no user story dependency)

---

## Notes

- `[P]` tasks operate on different files with no shared state — safe to run concurrently
- API layer (`internal/api/`) must always be complete and tested before its CLI counterpart (`internal/cmd/`)
- All mutation commands must output only the resource ID on stdout in non-TTY mode (verified in CLI tests)
- All error messages must match the contract in `contracts/cli-commands.md` exactly
- `go test ./...` must pass at every checkpoint before advancing to the next phase
- Never write API credentials to stdout or test logs
