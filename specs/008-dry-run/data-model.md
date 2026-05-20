# Data Model: Safe Exploration (--dry-run)

## New Type: DryRunOutput

Emitted to stdout (one JSON line) instead of making the HTTP request.

```go
// internal/api/client.go (or helpers.go)
type DryRunOutput struct {
    Action  string          `json:"action"`
    Target  string          `json:"target"`
    Payload json.RawMessage `json:"payload"`
}
```

### Field Definitions

| Field     | Type               | Description                                        | Example                    |
|-----------|--------------------|----------------------------------------------------|----------------------------|
| `action`  | `string`           | Semantic name of the operation                     | `"update"`, `"delete"`     |
| `target`  | `string`           | API path of the resource being acted on            | `"/admin/v1/clients/123"`  |
| `payload` | `json.RawMessage`  | JSON body that would have been sent; `null` if none | `{"alias":"acme-corp"}`   |

### Action Mapping

| HTTP Method     | `action` value |
|-----------------|----------------|
| `POST`          | `"create"`     |
| `PUT`           | `"update"`     |
| `PATCH`         | `"update"`     |
| `DELETE`        | `"delete"`     |
| `GET` (no-op)   | *(not emitted)*|

---

## Modified Type: api.Client

Add two fields to the existing `Client` struct in `internal/api/client.go`:

```go
type Client struct {
    BaseURL    *url.URL
    HTTPClient *http.Client
    UserAgent  string
    // New fields:
    DryRun bool      // when true, mutating requests are short-circuited
    Out    io.Writer // destination for dry-run JSON output (default: os.Stdout)
}
```

---

## Modified Function: api.NewClient

Add `dryRun bool` and `out io.Writer` parameters (appended to existing signature):

```go
// Before:
func NewClient(baseURL, token, version string, verbose bool, verboseOut io.Writer) (*Client, error)

// After:
func NewClient(baseURL, token, version string, verbose bool, verboseOut io.Writer, dryRun bool, out io.Writer) (*Client, error)
```

The constructor sets `Client.DryRun = dryRun` and `Client.Out = out`.

---

## Modified Function: api.doRequest

Add dry-run short-circuit before `c.HTTPClient.Do(req)`:

```go
// Pseudocode showing the interception point:
func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
    // ... existing: marshal body, parse URL, create req ...

    if c.DryRun && isMutatingMethod(method) {
        var rawPayload json.RawMessage
        if body != nil {
            rawPayload, _ = json.Marshal(body)
        } else {
            rawPayload = json.RawMessage("null")
        }
        out := DryRunOutput{
            Action:  httpMethodToAction(method),
            Target:  path,
            Payload: rawPayload,
        }
        _ = json.NewEncoder(c.Out).Encode(out)
        return &http.Response{
            StatusCode: http.StatusOK,
            Body:       io.NopCloser(bytes.NewReader(nil)),
        }, nil
    }

    // ... existing: c.HTTPClient.Do(req) ...
}
```

Helper functions (added in `helpers.go` or a small `dryrun.go`):

```go
func isMutatingMethod(method string) bool {
    switch strings.ToUpper(method) {
    case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
        return true
    }
    return false
}

func httpMethodToAction(method string) string {
    switch strings.ToUpper(method) {
    case http.MethodPost:
        return "create"
    case http.MethodPut, http.MethodPatch:
        return "update"
    case http.MethodDelete:
        return "delete"
    default:
        return strings.ToLower(method)
    }
}
```

---

## Root Command Changes

In `internal/cmd/root.go`, inside `NewRootCmd()`:

```go
var dryRunFlag bool

// In persistent flags section:
root.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Preview the action without executing it (prints JSON to stdout)")

// In PersistentPreRunE, when calling NewClient:
client, err := api.NewClient(cfg.BaseURL, cfg.Token, Version, verboseFlag, os.Stderr, dryRunFlag, os.Stdout)
```
