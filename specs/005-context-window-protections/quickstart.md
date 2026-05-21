# Quickstart: Context Window Protections

## Who this is for

AI agents and scripts that call `atadmin` list commands and need to limit the amount of data returned to avoid token budget overruns.

## The three tools

### 1. Field filtering — only return the keys you need

```sh
# Without protection: dumps every field for every user
atadmin users list --json

# With --fields: only id and email per user
atadmin users list --json --fields id,email

# Works on any list command
atadmin groups list --json --fields id,name
atadmin devices list --json --fields id,username,status
```

### 2. Safe pagination — automatic 50-item cap

No flag needed. When `--json` is used without an explicit `--limit`, the CLI caps results at 50.

```sh
# Returns at most 50 users (safe default)
atadmin users list --json

# Override to get more
atadmin users list --json --limit 200

# Override to get all (use with care in LLM context)
atadmin users list --json --limit 0
```

### 3. Summary mode — just the counts

```sh
# How many active users are there?
atadmin users list --json --summary --filter active30days
# → {"returned_items": 50, "has_more": true}

# Are there more groups than fit on one page?
atadmin groups list --json --summary
# → {"returned_items": 12, "has_more": false}
```

## Using with MCP (Claude Desktop / Cursor)

All three features are automatically available as MCP tool parameters — no configuration needed.

```
# In a conversation with Claude Desktop:
User: How many active users are there?
Claude: [calls users_list tool with {"summary": true, "filter": "active30days"}]
→ {"returned_items": 50, "has_more": true}

User: Show me just the IDs and emails of the first 10 users.
Claude: [calls users_list tool with {"limit": 10, "fields": "id,email"}]
→ [{"id": 1, "email": "a@example.com"}, ...]
```

## Combining flags

| Combination | Result |
|-------------|--------|
| `--json --fields id,email` | Filtered JSON, up to 50 items (safe default) |
| `--json --summary` | Aggregate stats only |
| `--json --summary --fields id` | Summary only (--fields ignored) |
| `--json --limit 100 --fields id,email` | Up to 100 items, filtered |
| Table mode (no --json) + `--fields` | `--fields` ignored; normal table output |

## Testing the feature

```sh
# Verify field filtering
atadmin users list --json --fields id,email | jq '.[0] | keys'
# Should output: ["email", "id"]

# Verify safe pagination
atadmin users list --json | jq length
# Should output: 50 (or fewer if account has fewer users)

# Verify summary
atadmin groups list --json --summary
# Should output: {"returned_items": N, "has_more": false/true}
```
