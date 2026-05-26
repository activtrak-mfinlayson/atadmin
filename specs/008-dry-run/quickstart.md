# Quickstart: --dry-run Flag

## What It Does

Running any mutating command with `--dry-run` prints a JSON preview to stdout instead of executing the operation. No HTTP request is sent.

## Usage

```bash
# Preview renaming a group
atadmin groups rename 42 "New Name" --dry-run

# Preview merging two clients
atadmin clients merge 100 200 --dry-run

# Preview a bulk alias update
atadmin clients alias bulk aliases.csv --dry-run

# Preview stopping tracking for a user
atadmin users bulk stop-tracking --ids user-1,user-2 --dry-run
```

## Output Format

Each dry-run prints one JSON object to stdout:

```json
{"action":"update","target":"/admin/v1/groups/42","payload":{"name":"New Name"}}
```

```json
{"action":"create","target":"/admin/v1/clients/mergeusers","payload":{"sourceUserId":100,"targetUserId":200}}
```

```json
{"action":"delete","target":"/admin/v1/clients/donottrack","payload":null}
```

## Parsing with jq

```bash
# Extract the target of a dry-run
atadmin groups rename 42 "New Name" --dry-run | jq .target

# Confirm action type before approving
atadmin users bulk delete-data --ids user-1 --dry-run | jq .action
```

## Piping to a Confirmation Step

```bash
# Human-in-the-loop pattern: review first, then execute
atadmin clients merge 100 200 --dry-run
# → {"action":"create","target":"/admin/v1/clients/mergeusers","payload":{...}}
# (review output, then run without --dry-run to apply)
atadmin clients merge 100 200
```

## Read-Only Commands Are Unaffected

`--dry-run` has no effect on `list`, `get`, or other read-only commands — they execute normally.

```bash
# These behave identically with or without --dry-run
atadmin clients list --dry-run
atadmin groups get 42 --dry-run
```

## MCP / Agent Usage

When used via the MCP server, agents can invoke any mutating tool with `dry_run: true` to generate a preview before requesting human approval.
