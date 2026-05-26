# Quickstart: Zero-Spillage Rule (Feature 006)

## What this feature fixes

Before this feature, `atadmin --json` output could be corrupted by two classes of "spillage":

1. **Mutation confirmation messages** going to stdout (e.g., `"Updated user 123\n"`)
2. **Plain-text errors** going to stderr even in JSON mode — agents can't parse them

After this feature:
- `atadmin <any-command> --json` → stdout contains **exactly one** valid JSON object or array
- A failed `atadmin <any-command> --json` → stdout contains `{"error": "...", "suggestion": "..."}`
- All confirmation/diagnostic text → stderr, regardless of mode

---

## How to verify the fix

### 1. Stdout purity (read command)

```sh
# Before: always worked — read commands already output JSON
atadmin users list --json | jq length

# After: still works, and guaranteed clean
atadmin users list --json > /tmp/users.json
jq length /tmp/users.json   # no parse errors
```

### 2. Stdout purity (mutation command)

```sh
# Before: "Updated user 123\n" would appear in stdout, breaking pipes
atadmin users update 123 --display-name "Foo" --json > /dev/null
# After: confirmation goes to stderr; stdout is empty (or JSON if applicable)
atadmin users update 123 --display-name "Foo" 2>/dev/null   # stderr suppressed
echo "exit: $?"                                               # 0 = success
```

### 3. Structured error (JSON mode)

```sh
# With an invalid/expired token:
atadmin users list --json
# stdout:
# {
#   "error": "users list: 401 Unauthorized",
#   "suggestion": "Run 'atadmin auth login' to authenticate."
# }
# exit code: 1

# Pipe-safe error handling in a script:
result=$(atadmin users list --json 2>/dev/null)
if echo "$result" | jq -e '.error' > /dev/null 2>&1; then
  echo "Error: $(echo "$result" | jq -r '.error')"
  echo "Hint:  $(echo "$result" | jq -r '.suggestion // empty')"
fi
```

### 4. Non-JSON mode unchanged

```sh
# Table output still works exactly as before
atadmin users list

# Errors still go to stderr in non-JSON mode
atadmin users list 2>&1 | grep "Error:"
```

---

## Key implementation locations

| File | Change |
|---|---|
| `internal/output/output.go` | New: `WriteError()`, `DetectJSONMode()`, `SuggestionFor()` |
| `internal/cmd/root.go` | Modified: `Execute()` uses `WriteError` + `DetectJSONMode` |
| `internal/cmd/users.go` | ~6 confirmation messages → `ErrOrStderr` |
| `internal/cmd/groups.go` | ~10 confirmation messages → `ErrOrStderr` |
| `internal/cmd/consumers.go` | ~5 confirmation messages → `ErrOrStderr` |

---

## For agent integration

Agents calling `atadmin --json` should:

1. Always capture stdout (the JSON payload)
2. Always discard or log stderr (diagnostic/human messages)
3. Check if stdout contains `{"error": ...}` before trusting the data
4. Use the `suggestion` field to self-correct automatically

```python
import subprocess, json

result = subprocess.run(
    ["atadmin", "users", "list", "--json"],
    capture_output=True, text=True
)
data = json.loads(result.stdout)
if "error" in data:
    raise RuntimeError(f"{data['error']} — {data.get('suggestion', '')}")
users = data  # list of user objects
```
