# Quickstart: Embedded MCP Server

**Feature**: 004-mcp-server

---

## Scenario 1: Start the MCP Server Manually

```bash
# Ensure atadmin is authenticated
atadmin auth login

# Start the MCP server on stdio
atadmin mcp serve --stdio
```

The server blocks, waiting for JSON-RPC messages on stdin.

---

## Scenario 2: Connect Claude Desktop

1. Open `~/Library/Application Support/Claude/claude_desktop_config.json`
2. Add the `atadmin` server entry:

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

3. Restart Claude Desktop.
4. In a new conversation, ask: _"List the first 5 users in my ActivTrak account."_
5. Claude automatically invokes `users_list` with `{"limit": 5, "json": true}` and presents the results.

---

## Scenario 3: Discover Available Tools (via curl/manual test)

Send a `tools/list` request to verify the server works:

```bash
# Start server in background with a named pipe
mkfifo /tmp/mcp-in /tmp/mcp-out
atadmin mcp serve --stdio < /tmp/mcp-in > /tmp/mcp-out &

# Initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' > /tmp/mcp-in
echo '{"jsonrpc":"2.0","id":2,"method":"notifications/initialized","params":{}}' > /tmp/mcp-in

# List tools
echo '{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}' > /tmp/mcp-in
cat /tmp/mcp-out
```

Expected: JSON response with 20+ tools listed.

---

## Scenario 4: Call a Tool Directly

```bash
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"users_list","arguments":{"limit":3,"json":true}}}' > /tmp/mcp-in
cat /tmp/mcp-out
```

Expected: JSON response with `content[0].text` containing the raw JSON from the ActivTrak API.

---

## Scenario 5: Use a Non-Default Profile

When calling from an agent, pass `profile` as a parameter to target a specific account:

```json
{
  "name": "users_list",
  "arguments": {
    "profile": "staging",
    "limit": 10,
    "json": true
  }
}
```

This is equivalent to running `atadmin --profile staging users list --limit 10 --json`.

---

## Scenario 6: Error Handling

If the API token is expired, any tool call returns:

```json
{
  "result": {
    "content": [{ "type": "text", "text": "listing users: unauthorized. Run 'atadmin auth login'." }],
    "isError": true
  }
}
```

The server stays alive; the agent can prompt the user to re-authenticate and retry.
