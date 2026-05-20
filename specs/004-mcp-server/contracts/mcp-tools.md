# Contract: MCP Tool Interface

**Feature**: 004-mcp-server  
**Date**: 2026-05-19  
**Protocol**: Model Context Protocol (MCP) 2024-11-05+

This document specifies the MCP interface that `atadmin mcp serve --stdio` exposes to clients (Claude Desktop, Cursor, Craft Agent, etc.).

---

## Transport

- **Mode**: stdio (stdin/stdout)
- **Encoding**: JSON-RPC 2.0, newline-delimited
- **Start command**: `atadmin mcp serve --stdio`
- **Read-only mode (default)**: Only non-mutating commands are exposed as tools
- **Mutation mode**: Pass `--allow-mutations` to also expose commands classified as mutations (those named `update`, `delete`, `remove`, `bulk`, `add`, `create`)
- **Diagnostic log**: `~/.config/atadmin/mcp.log` — all startup and tool-call events written here; nothing written to stdout or stderr after the server starts

---

## Initialize Handshake

### Client → Server

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": { "name": "claude-desktop", "version": "1.0.0" }
  }
}
```

### Server → Client

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "atadmin",
      "version": "0.1.0"
    }
  }
}
```

---

## Tool Discovery: `tools/list`

### Request

```json
{ "jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {} }
```

### Response

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "users_list",
        "description": "List identity entities (users) in the account",
        "inputSchema": {
          "type": "object",
          "properties": {
            "filter": {
              "type": "string",
              "description": "Filter type (tracked, untracked, usersWithAgents, ...)"
            },
            "search": {
              "type": "string",
              "description": "Search term"
            },
            "limit": {
              "type": "integer",
              "description": "Maximum results (0 = server default)"
            },
            "cursor": {
              "type": "string",
              "description": "Pagination cursor"
            },
            "json": {
              "type": "boolean",
              "description": "Output raw JSON"
            },
            "profile": {
              "type": "string",
              "description": "Config profile to use (env: ATADMIN_PROFILE)"
            },
            "format": {
              "type": "string",
              "description": "Output format: table or json (env: ATADMIN_FORMAT)"
            }
          },
          "required": []
        }
      },
      {
        "name": "users_get",
        "description": "Get a single identity entity by ID",
        "inputSchema": {
          "type": "object",
          "properties": {
            "id": {
              "type": "string",
              "description": "Entity ID"
            },
            "json": {
              "type": "boolean",
              "description": "Output raw JSON"
            }
          },
          "required": []
        }
      }
    ]
  }
}
```

> **Security note**: The `--token` and `--base-url` flags are intentionally excluded from all tool parameter schemas. Auth credentials must be pre-configured via `atadmin auth login` before starting the MCP server. This prevents API tokens from appearing in agent request logs or conversation history.

**Tool naming convention**: `{resource}_{action}` using underscores. Examples:
- `atadmin users list` → `users_list`
- `atadmin users get` → `users_get`
- `atadmin groups list` → `groups_list`
- `atadmin audit-log list` → `audit_log_list`
- `atadmin api-keys list` → `api_keys_list`

---

## Tool Execution: `tools/call`

### Request

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "users_list",
    "arguments": {
      "limit": 10,
      "json": true
    }
  }
}
```

### Successful Response (JSON output)

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"results\":[{\"id\":12345,\"displayName\":{\"value\":\"Alice Smith\"}}],\"cursor\":\"\",\"totalCount\":1}"
      }
    ],
    "isError": false
  }
}
```

### Successful Response (table output, no `--json` flag)

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "ID      STATUS   NAME\n12345   active   Alice Smith\n"
      }
    ],
    "isError": false
  }
}
```

### Error Response (command failure)

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "listing users: unauthorized. Your token may have expired. Try running 'atadmin auth login'."
      }
    ],
    "isError": true
  }
}
```

### Error Response (unknown tool)

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "error": {
    "code": -32601,
    "message": "tool not found: nonexistent_tool"
  }
}
```

---

## Parameter → CLI Flag Mapping

When the server receives a `tools/call`, it reconstructs the CLI arguments:

| Parameter type | Encoding rule                                              |
|----------------|------------------------------------------------------------|
| `boolean: true` | Append `--{flag-name}` (no value)                        |
| `boolean: false` | Omit the flag                                            |
| `integer: N`    | Append `--{flag-name}`, `"{N}"`                          |
| `string: "v"`   | Append `--{flag-name}`, `"{v}"`                          |
| Positional args | Appended after flags (for commands like `users get <id>`) |

---

## Claude Desktop Configuration

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "atadmin": {
      "command": "atadmin",
      "args": ["mcp", "serve", "--stdio"]
    }
  }
}
```

If `atadmin` is not on PATH, use the absolute binary path:

```json
{
  "mcpServers": {
    "atadmin": {
      "command": "/usr/local/bin/atadmin",
      "args": ["mcp", "serve", "--stdio"]
    }
  }
}
```

---

## Cursor Configuration

Add to `.cursor/mcp.json` in the workspace or `~/.cursor/mcp.json` globally:

```json
{
  "mcpServers": {
    "atadmin": {
      "command": "atadmin",
      "args": ["mcp", "serve", "--stdio"]
    }
  }
}
```
