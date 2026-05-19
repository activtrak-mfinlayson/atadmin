# CLI Command Contract: atadmin

All commands follow the noun-verb structure: `atadmin <resource> <action> [flags]`

## Global Flags (all commands)

| Flag | Short | Default | Description |
|---|---|---|---|
| `--profile` | | `default` | Named config profile to use |
| `--format` | `-f` | `table` | Output format: `table` or `json` |
| `--verbose` | | false | Print HTTP method, URL, status, retries to stderr |
| `--timeout` | | `30s` | Per-request timeout |
| `--token` | | | Override bearer token (env: `ATADMIN_TOKEN`) |
| `--base-url` | | | Override API base URL (env: `ATADMIN_BASE_URL`) |

---

## auth

```
atadmin auth login [--profile <name>]
```

Prompts user (masked input) for a bearer token, validates it against the API ping endpoint, and saves it to the named profile in `~/.config/atadmin/config.yaml` with `0600` permissions.

| Subcommand | Description |
|---|---|
| `login` | Store credentials for a profile |

---

## clients

```
atadmin clients <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/clients` | `--page`, `--page-size`, `--json` |
| `get <id\|username>` | `GET /admin/v1/clients/{clientId}` or `/{username}` | |
| `update <id>` | `PUT /admin/v1/clients/{clientId}` | `--alias <str>` |
| `delete` | `DELETE /admin/v1/clients` | `--ids <id,...>` |
| `restore` | `PUT /admin/v1/clients/restore` | `--ids <id,...>` |
| `merge` | `POST /admin/v1/clients/mergeusers` | `--source <id>`, `--target <id>` |
| `merge bulk` | `POST /admin/v1/clients/mergeusers/bulk` | `--file <path>` |
| `unmerge bulk` | `DELETE /admin/v1/clients/unmergeusers/bulk` | `--file <path>` |
| `alias set` | `PUT /admin/v1/clients/useralias` | `--id <id>`, `--alias <str>` |
| `alias bulk` | `POST /admin/v1/clients/useralias/bulk` | `--file <path>` |
| `donottrack list` | `GET /admin/v1/clients/donottrack` | |
| `donottrack add` | `POST /admin/v1/clients/donottrack` | `--domain <str>`, `--username <str>` |
| `donottrack remove` | `DELETE /admin/v1/clients/donottrack` | `--ids <id,...>` |
| `donottrack update` | `PUT /admin/v1/clients/donottrack` | `--id <id>`, `--domain <str>`, `--username <str>` |
| `donottrack add-bulk` | `POST /admin/v1/clients/donottrack/bulk` | `--file <path>` |
| `donottrack remove-bulk` | `DELETE /admin/v1/clients/donottrack/bulk` | `--file <path>` |
| `donottrack global-user` | `PATCH /admin/v1/clients/donottrack/globaluser` | `--ids <id,...>` |
| `health` | `GET /admin/v1/clients/health` | |

**Output**: `list` → table (username, alias, status); `get` → key-value; mutations → ID on stdout

---

## groups

```
atadmin groups <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/groups/list` | `--page`, `--page-size` |
| `summary` | `GET /admin/v1/groups/summary` | |
| `get <id>` | `GET /admin/v1/groups/list/{id}` | |
| `search <prefix>` | `GET /admin/v1/groups/list/{prefix}` | |
| `create <name>` | `POST /admin/v1/groups/{name}` | |
| `rename <id>` | `PUT /admin/v1/groups/{groupId}` | `--name <str>` |
| `delete` | `DELETE /admin/v1/groups` | `--ids <id,...>` |
| `members list` | `GET /admin/v1/groups/members` | `--page`, `--page-size` |
| `members list <group-id>` | `GET /admin/v1/groups/{groupId}/members` | |
| `members add` | `POST /admin/v1/groups/members` | `--group <id>`, `--member <id>`, `--type client\|device` |
| `members remove` | `DELETE /admin/v1/groups/members` | `--group <id>`, `--member <id>` |
| `members export` | `GET /admin/v1/groups/members/bulk` | `--output <path>` |
| `members import` | `POST /admin/v1/groups/members/bulk` | `--file <path>` |
| `clients add <group-id>` | `POST /admin/v1/groups/{groupId}/clients` | `--ids <id,...>` |
| `clients remove <group-id>` | `DELETE /admin/v1/groups/{groupId}/clients` | `--ids <id,...>` |
| `devices add <group-id>` | `POST /admin/v1/groups/{groupId}/devices` | `--ids <id,...>` |
| `devices remove <group-id>` | `DELETE /admin/v1/groups/{groupId}/devices` | `--ids <id,...>` |
| `membership get <group-id> <type> <member-id>` | `GET /admin/v1/groups/{groupId}/members/{memberType}/{memberId}` | |

---

## consumers

```
atadmin consumers <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/consumers` | `--page`, `--page-size` |
| `get <id>` | `GET /admin/v1/consumers/{id}` | |
| `create` | `POST /admin/v1/consumers` | `--file <path>` or inline flags |
| `update` | `PATCH /admin/v1/consumers` | `--file <path>` |
| `delete` | `DELETE /admin/v1/consumers` | `--ids <id,...>` |
| `delete bulk` | `DELETE /admin/v1/consumers/bulk` | `--file <path>` |
| `role set <id>` | `PUT /admin/v1/consumers/{consumerId}/role` | `--role <str>` |
| `password set <id>` | `PUT /admin/v1/consumers/{consumerId}/password` | (prompts interactively) |
| `sso set` | `PUT /admin/v1/consumers/usesso` | `--consumer <id>`, `--use-sso <bool>` |
| `groups add <id>` | `PATCH /admin/v1/consumers/{consumerId}/viewablegroups` | `--group-ids <id,...>` |
| `chrome-users import` | `POST /admin/v1/consumers/chromeusers/bulk` | `--file <path>` |

---

## devices

```
atadmin devices <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/devices` | `--page`, `--page-size` |
| `get <id>` | `GET /admin/v1/devices/{deviceId}` | |
| `delete` | `DELETE /admin/v1/devices` | `--ids <id,...>` |
| `restore` | `PUT /admin/v1/devices/restore` | `--ids <id,...>` |
| `uninstall` | `POST /admin/v1/devices/uninstall` | `--ids <id,...>` |

---

## settings

```
atadmin settings <sub-resource> <get|set> [flags]
```

All `get` commands: key-value output. All `set` commands: success confirmation on stdout.

| Subcommand | API |
|---|---|
| `privacy get` / `privacy set` | `GET/PUT /admin/v1/accountsettings/privacy` |
| `sso get` / `sso set` | `GET/PUT /admin/v1/accountsettings/sso` |
| `sso enabled` | `GET /admin/v1/accountsettings/sso/enabled` |
| `sso eligible` | `GET /admin/v1/accountsettings/sso/eligible` |
| `role-access get` / `role-access set` | `GET/POST /admin/v1/accountsettings/roleaccess` |
| `role-access reset` | `POST /admin/v1/accountsettings/roleaccess/reset` |
| `role-date-filter get` / `role-date-filter set` | `GET/POST /admin/v1/accountsettings/roledatefilter` |
| `timezone get` / `timezone set` | `GET/POST /admin/v1/accountsettings/timezone` |
| `timezones list` | `GET /admin/v1/accountsettings/timezones` |
| `local-timezone get` / `local-timezone set` | `GET/POST /admin/v1/accountsettings/showlocaltimezone` |
| `agent-duration get` / `agent-duration set` / `agent-duration delete` | `GET/PUT/DELETE /admin/v1/accountsettings/agent/activityduration` |
| `agent-audit get` / `agent-audit set` | `GET/POST /admin/v1/accountsettings/agent/audit` |
| `passive-time get` / `passive-time set` | `GET/PATCH /admin/v1/accountsettings/computerpassivetime` |
| `passive-time bulk-set` | `POST /admin/v1/accountsettings/computerpassivetime/bulk` |
| `schedule-adherence get` / `schedule-adherence set` | `GET/PUT /admin/v1/accountsettings/schedule_adherence` |
| `email-autodetect get` / `email-autodetect set` | `GET/POST /admin/v1/accountsettings/emailautodetection` |
| `identity-match get` / `identity-match set` | `GET/POST /admin/v1/accountsettings/identitynewagentmatchuser` |
| `identity-threshold get` / `identity-threshold set` | `GET/POST /admin/v1/accountsettings/identitysearchactivethresholddays` |
| `license-approval get` / `license-approval set` | `GET/POST /admin/v1/accountsettings/licenseapprovalmode` |
| `msp-overage get` / `msp-overage set` / `msp-overage delete` | `GET/PUT/DELETE /admin/v1/accountsettings/msp_license_overage` |
| `hris get` | `GET /admin/v1/accountsettings/hris` |
| `academy url` | `GET /admin/v1/accountsettings/academy` |
| `academy workramp-url` | `GET /admin/v1/accountsettings/academy/workramp` |
| `ping` | `GET /admin/v1/accounts/ping` |

---

## alarms

```
atadmin alarms <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/alarms` | `--page`, `--page-size` |
| `get <id>` | `GET /admin/v1/alarms/{id}` | |
| `details <id>` | `GET /admin/v1/alarmdetails/{id}` | |
| `create` | `POST /admin/v1/alarms` | `--file <path>` |
| `update` | `PUT /admin/v1/alarms` | `--file <path>` |
| `delete <id>` | `DELETE /admin/v1/alarms/{id}` | |
| `conditions` | `GET /admin/v1/alarms/conditions` | |
| `fields` | `GET /admin/v1/alarms/fields` | |

---

## signals

```
atadmin signals <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/signals` | |
| `create` | `POST /admin/v1/signal` | `--file <path>` |
| `update` | `PUT /admin/v1/signal` | `--file <path>` |
| `delete <id>` | `DELETE /admin/v1/signals/{id}` | |

---

## schedules

```
atadmin schedules <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/schedules` | |
| `get <id>` | `GET /admin/v1/schedules/{id}` | |
| `create` | `POST /admin/v1/schedule` | `--file <path>` |
| `delete <id>` | `DELETE /admin/v1/schedules/{id}` | |
| `reporting default get` | `GET /admin/v1/schedules/reporting/default` | |
| `reporting default set <id>` | `PUT /admin/v1/schedules/reporting/default/{scheduleId}` | |
| `reporting users list` | `GET /admin/v1/schedules/reporting/users` | |
| `reporting users remove` | `DELETE /admin/v1/schedules/reporting/users` | `--ids <id,...>` |
| `shift default get` | `GET /admin/v1/schedules/shift/default` | |
| `shift default set <id>` | `PUT /admin/v1/schedules/shift/default/{scheduleId}` | |
| `shift users list` | `GET /admin/v1/schedules/shift/users` | |
| `shift users remove` | `DELETE /admin/v1/schedules/shift/users` | `--ids <id,...>` |
| `users list <id>` | `GET /admin/v1/schedules/{scheduleId}/users` | |
| `users set <id>` | `PUT /admin/v1/schedules/{scheduleId}/users` | `--ids <id,...>` |
| `user move` | `PUT /admin/v1/schedules/{scheduleId}/user/{userId}` | `--schedule <id>`, `--user <id>` |
| `user get <user-id>` reporting | `GET /admin/v1/user/{userId}/schedule/reporting` | |
| `user get <user-id>` shift | `GET /admin/v1/user/{userId}/schedule/shift` | |
| `user remove <user-id>` reporting | `DELETE /admin/v1/user/{userId}/schedule/reporting` | |
| `user remove <user-id>` shift | `DELETE /admin/v1/user/{userId}/schedule/shift` | |

---

## apikeys

```
atadmin apikeys <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/key` | |
| `create` | `POST /admin/v1/key` | `--name <str>` |
| `update` | `PUT /admin/v1/key` | `--id <id>`, `--name <str>` |
| `delete <id>` | `DELETE /admin/v1/key/{keyId}` | |
| `util backfill all` | `POST /admin/v1/util/backfill_apikey_instanceid` | |
| `util backfill <id>` | `POST /admin/v1/util/backfill_apikey_instanceid/{apiKeyId}` | |

---

## auditlog

```
atadmin auditlog <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `list` | `GET /admin/v1/auditlog` | `--from <datetime>`, `--to <datetime>`, `--filters <str>`, `--sort <col>`, `--desc`, `--page`, `--page-size` |
| `attachment get <id>` | `GET /admin/v1/attachment/{attachmentid}` | |

---

## hrdc

```
atadmin hrdc <action> [flags]
```

| Subcommand | API | Key Flags |
|---|---|---|
| `ping` | `GET /hrdc/ping` | |
| `import` | `POST /hrdc/v1/bulk` | `--file <path>` (.json or .csv) |

---

## notifications (deprecated)

```
atadmin notifications list   [DEPRECATED]
```

| Subcommand | API |
|---|---|
| `list` `[DEPRECATED]` | `GET /admin/legacy/notifications` |

Alias for `atadmin signals list` using the legacy endpoint. Warns on stderr when used.

---

## Error Message Contract

All API errors are mapped to actionable messages:

| HTTP Status | Message Pattern |
|---|---|
| 400 | `Error: Bad request — <api message>. Check your flags and try again.` |
| 401 | `Error: Unauthorized. Your token may have expired. Try running 'atadmin auth login'.` |
| 403 | `Error: Forbidden. Your account role may not have permission for this operation.` |
| 404 | `Error: Not found. The requested <resource> does not exist.` |
| 429 | `Error: Rate limited. Retried <N> times. Wait a moment and try again.` |
| 5xx | `Error: Server error (<status>). The ActivTrak API returned an unexpected error. Try again later.` |
| timeout | `Error: Request timed out after <duration>. Check your network or increase --timeout.` |
| connection | `Error: Could not connect to <url>. Check your network or --base-url setting.` |
