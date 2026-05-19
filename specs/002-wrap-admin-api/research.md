# Research: ActivTrak Admin API CLI Wrapper

## Named Configuration Profiles (Viper)

**Decision**: Nested YAML under `profiles.<name>.*`; loaded with `v.Sub("profiles."+name)` into a sub-Viper instance.

**Rationale**: Viper's `Sub()` method returns a new Viper scoped to a key prefix, which gives clean isolation per profile without a separate config file per profile. The active profile is selected by the global `--profile` flag (default: `"default"`). The `ATADMIN_TOKEN` environment variable overrides the profile token when set, maintaining env-var-override semantics.

Config file shape:
```yaml
profiles:
  default:
    token: "abc123"
    base_url: "https://api.activtrak.com"
  staging:
    token: "xyz789"
    base_url: "https://staging-api.activtrak.com"
```

**Alternatives considered**:
- Separate `profiles.d/staging.yaml` files — adds filesystem complexity, harder to manage
- Flat YAML with prefixed keys (`staging_token`) — doesn't scale to N profiles

---

## Retry on HTTP 429 (stdlib-only)

**Decision**: Implement `retryRoundTripper` wrapping the inner transport with exponential backoff (`2^attempt` seconds, max 3 retries, 4 total attempts).

**Rationale**: Standard library `time` and `math` are sufficient. No new dependency is needed. This aligns with CLAUDE.md preference for avoiding bloated HTTP frameworks. The chain is: `verboseRoundTripper → retryRoundTripper → authRoundTripper → http.DefaultTransport`.

```go
type retryRoundTripper struct {
    inner    http.RoundTripper
    maxRetry int
}

func (r *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    for attempt := 0; attempt <= r.maxRetry; attempt++ {
        resp, err := r.inner.RoundTrip(req)
        if err != nil || resp.StatusCode != http.StatusTooManyRequests {
            return resp, err
        }
        if attempt < r.maxRetry {
            time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
        }
    }
    // resp is 429 after max retries
    return resp, nil
}
```

**Alternatives considered**: `github.com/cenkalti/backoff` — feature-rich but adds a dependency; hardcoded sleep intervals — fragile.

---

## TTY Detection

**Decision**: Use `golang.org/x/term` — `term.IsTerminal(int(os.Stdout.Fd()))`.

**Rationale**: This is explicitly prescribed in CLAUDE.md. The one-liner detects whether stdout is connected to a terminal, enabling color/table output in interactive mode and bare output in piped/CI mode. Add with `go get golang.org/x/term`.

---

## Table Output (text/tabwriter)

**Decision**: Use `text/tabwriter` from the standard library.

**Rationale**: CLAUDE.md explicitly prescribes `text/tabwriter`. Pattern: create writer → `fmt.Fprintf` tab-delimited rows → `Flush()`.

```go
w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
fmt.Fprintln(w, "NAME\tSTATUS\tALIAS")
for _, c := range clients {
    fmt.Fprintf(w, "%s\t%s\t%s\n", c.Username, c.Status, c.Alias)
}
w.Flush()
```

---

## Bulk File Parsing (JSON + CSV)

**Decision**: Auto-detect by `filepath.Ext()`, parse with `encoding/json` (JSON) and `encoding/csv` (CSV). Both are standard library — zero new dependencies.

**Rationale**: CSV is header-row based, resulting in `[]map[string]string`; JSON results in `[]map[string]any`. The bulk API methods receive a typed slice, so the parsing layer must map the raw maps to the appropriate request struct for each resource.

```go
ext := strings.ToLower(filepath.Ext(path))
switch ext {
case ".json":
    var records []map[string]any
    json.Unmarshal(data, &records)
case ".csv":
    r := csv.NewReader(file)
    headers, _ := r.Read()
    for { row, err := r.Read(); /* map headers to row */ }
default:
    return fmt.Errorf("unsupported file format %q; expected .json or .csv", ext)
}
```

---

## Key-Value Single-Object Output

**Decision**: Print `key: value` pairs using left-padded formatting via `fmt.Fprintf`. No library needed.

**Rationale**: Simple, consistent with git config output style, works at any terminal width.

```go
func PrintKeyValue(out io.Writer, fields map[string]string) {
    keys := sortedKeys(fields)
    for _, k := range keys {
        fmt.Fprintf(out, "%-24s %s\n", k+":", fields[k])
    }
}
```
