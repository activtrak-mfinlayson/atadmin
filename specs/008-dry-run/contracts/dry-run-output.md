# Contract: --dry-run Output Schema

## Trigger

A mutating command (`POST`, `PUT`, `PATCH`, `DELETE`) is invoked with the `--dry-run` flag.

## Output

One JSON object is written to **stdout** per HTTP request that would have been made. The output is a single line (no pretty-printing) for machine parseability.

### Schema

```json
{
  "action":  "<string>",
  "target":  "<string>",
  "payload": <object|null>
}
```

| Field     | Type          | Required | Description                                                       |
|-----------|---------------|----------|-------------------------------------------------------------------|
| `action`  | `string`      | yes      | Operation type: `"create"`, `"update"`, or `"delete"`            |
| `target`  | `string`      | yes      | API path of the resource (e.g., `/admin/v1/clients/123`)         |
| `payload` | object\|null  | yes      | JSON body that would have been sent; `null` if no body            |

### Action Values

| Value      | Triggered by HTTP method |
|------------|--------------------------|
| `"create"` | `POST`                   |
| `"update"` | `PUT` or `PATCH`         |
| `"delete"` | `DELETE`                 |

## Guarantees

1. **No side effects**: When `--dry-run` is active, zero HTTP requests are made to the remote API for mutating operations.
2. **Exit code 0**: A successful dry-run exits with code `0`.
3. **Read commands unaffected**: `GET` requests are never intercepted; they execute normally.
4. **Machine-readable**: Output is valid JSON, one object per line.
5. **stderr is clean**: No diagnostic output is written to stderr by dry-run itself (verbose output from `--verbose` still appears if that flag is also set).

## Examples

```
$ atadmin groups rename 42 "Engineering" --dry-run
{"action":"update","target":"/admin/v1/groups/42","payload":{"name":"Engineering"}}

$ atadmin clients merge 100 200 --dry-run
{"action":"create","target":"/admin/v1/clients/mergeusers","payload":{"sourceUserId":100,"targetUserId":200}}

$ atadmin devices uninstall 55 --dry-run
{"action":"create","target":"/admin/v1/devices/uninstall","payload":{"ids":[55]}}

$ atadmin consumers delete 77 --dry-run
{"action":"delete","target":"/admin/v1/consumers","payload":null}
```

## Error Cases

| Scenario                       | Behaviour                                         |
|--------------------------------|---------------------------------------------------|
| Flag passed to a read command  | Command executes normally; no dry-run output      |
| Invalid arguments (pre-flight) | Command returns error before dry-run is reached   |
| Auth token missing             | Error returned (dry-run check is after auth setup)|
