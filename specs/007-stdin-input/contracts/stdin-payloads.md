# Contract: `--from-stdin` Payload Schemas

All commands that accept `--from-stdin` expect **JSON** on stdin. Other formats are out of scope.

---

## Group B: Typed-Struct Commands

### `users update <id> --from-stdin`

Replaces all individual flags. Fields map directly to `api.UpdateUserRequest`. Absent fields are not patched.

```json
{
  "displayName": "Jane Smith",
  "firstName": "Jane",
  "lastName": "Smith",
  "timezone": "America/Chicago",
  "tracked": true
}
```

All fields are optional. At least one must be present (validation identical to flag path).

---

### `users bulk start-tracking --from-stdin`
### `users bulk stop-tracking --from-stdin`
### `users bulk delete-entity --from-stdin`
### `users bulk delete-data --from-stdin`

Replaces `--ids`. The caller supplies the full `api.BulkActionRequest`.

```json
{
  "actions": ["StartTracking"],
  "data": [
    {"entityId": 101, "revision": 3},
    {"entityId": 102, "revision": 7}
  ]
}
```

The `actions` array must match the subcommand's expected action string. The command validates this and returns an error on mismatch.

---

## Group C: Confirmation-Bypass Commands

### `users delete <id> --from-stdin`

No payload read from stdin. The flag solely bypasses the interactive y/N confirmation prompt. Equivalent to `--yes`. Stdin must not contain data (or data is ignored).

---

### `consumers password set <id> --from-stdin`

Reads a single JSON object with a `password` field. Replaces the interactive masked-input prompt.

```json
{
  "password": "s3cur3P@ssword!"
}
```

`password` is required. Empty string returns an error.

---

## Group A: File-Record Commands

All file-record commands that currently accept `--file path.json` also accept `--from-stdin` as an alternative. The JSON format on stdin is identical to what the file would contain.

`--file` and `--from-stdin` are mutually exclusive; providing both returns an error.

### `clients merge-bulk --from-stdin`

```json
[
  {"sourceId": 10, "targetId": 20},
  {"sourceId": 11, "targetId": 21}
]
```

### `clients unmerge-bulk --from-stdin`

```json
[
  {"sourceId": 10},
  {"sourceId": 11}
]
```

### `clients alias-bulk --from-stdin`

```json
[
  {"id": 10, "alias": "alice"},
  {"id": 11, "alias": "bob"}
]
```

### `clients donottrack add-bulk --from-stdin`

```json
[
  {"logondomain": "CORP", "username": "jdoe"},
  {"logondomain": "CORP", "username": "asmith"}
]
```

### `clients donottrack remove-bulk --from-stdin`

```json
[
  {"id": 5},
  {"id": 6}
]
```

### `consumers delete-bulk --from-stdin`

```json
[
  {"id": 100},
  {"id": 101}
]
```

### `consumers chrome-users import --from-stdin`

```json
[
  {"email": "user@example.com", "role": "viewer"},
  {"email": "admin@example.com", "role": "admin"}
]
```

### `consumers create --from-stdin` / `consumers update --from-stdin`

Follows the same schema as the corresponding `--file` format.

### `groups members import --from-stdin`

```json
[
  {"memberId": 200, "memberType": "user"},
  {"memberId": 201, "memberType": "user"}
]
```

### `hrdc import --from-stdin`

```json
[
  {"employeeId": "E001", "email": "user@corp.com"},
  {"employeeId": "E002", "email": "user2@corp.com"}
]
```

### `schedules create --from-stdin`

Follows the existing schedule JSON structure (same as `--file`).

### `signals create --from-stdin` / `signals update --from-stdin`

Follows the existing signal JSON structure (same as `--file`).

### `alarms create --from-stdin` / `alarms update --from-stdin`

Follows the existing alarm JSON structure (same as `--file`).

---

## Error Responses

All `--from-stdin` errors are written to `stderr` in the format:

```
Error: --from-stdin: <reason>
```

Common reasons:
- `stdin is empty; pipe a JSON payload`
- `invalid JSON: <json.SyntaxError details>`
- `"<field>" field is required`
- `--file and --from-stdin are mutually exclusive`
