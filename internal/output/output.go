package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

// JSONError is the structured error emitted to stdout when --json is active and a command fails.
type JSONError struct {
	Error      string `json:"error"`
	Suggestion string `json:"suggestion,omitempty"`
}

// WriteError writes err to out in the appropriate format.
// When asJSON is true it emits a JSONError object; otherwise it writes "Error: <msg>\n".
func WriteError(out io.Writer, err error, suggestion string, asJSON bool) {
	if asJSON {
		je := JSONError{Error: err.Error(), Suggestion: suggestion}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(je)
		return
	}
	_, _ = fmt.Fprintf(out, "Error: %s\n", err)
}

// DetectJSONMode scans args for any flag pattern that signals JSON output mode.
func DetectJSONMode(args []string) bool {
	for i, arg := range args {
		if arg == "--json" {
			return true
		}
		if arg == "--format=json" || arg == "-f=json" {
			return true
		}
		if (arg == "--format" || arg == "-f") && i+1 < len(args) && args[i+1] == "json" {
			return true
		}
	}
	return false
}

// SuggestionFor derives an actionable suggestion from err by inspecting its message.
func SuggestionFor(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "loading profile") {
		return "Run 'atadmin auth login' to configure credentials."
	}
	if strings.Contains(msg, "401") || strings.Contains(strings.ToLower(msg), "unauthorized") {
		return "Run 'atadmin auth login' to authenticate."
	}
	if strings.Contains(msg, "404") || strings.Contains(strings.ToLower(msg), "not found") {
		return "Check the resource ID and try again."
	}
	if strings.Contains(msg, "500") || strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") || strings.Contains(msg, "504") ||
		strings.Contains(strings.ToLower(msg), "server error") {
		return "The ActivTrak API encountered an error. Try again later."
	}
	return ""
}

// Table writes a tab-aligned table to out with the given headers and rows.
func Table(out io.Writer, headers []string, rows [][]string) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			_, _ = fmt.Fprint(w, "\t")
		}
		_, _ = fmt.Fprint(w, h)
	}
	_, _ = fmt.Fprintln(w)
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				_, _ = fmt.Fprint(w, "\t")
			}
			_, _ = fmt.Fprint(w, cell)
		}
		_, _ = fmt.Fprintln(w)
	}
	_ = w.Flush()
}

// KeyValue writes key: value pairs to out, one per line, sorted by key.
func KeyValue(out io.Writer, fields map[string]string) {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, _ = fmt.Fprintf(out, "%-28s %s\n", k+":", fields[k])
	}
}

// JSON marshals v as indented JSON and writes it to out.
func JSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
