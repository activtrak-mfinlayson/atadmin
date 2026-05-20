package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

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

// FilterFields returns data with only the specified top-level keys retained.
// Handles map[string]any (single object) and slices thereof (array of objects).
// Other types are returned unchanged.
func FilterFields(data any, fields []string) any {
	allowed := make(map[string]bool, len(fields))
	for _, f := range fields {
		allowed[strings.TrimSpace(f)] = true
	}
	switch v := data.(type) {
	case map[string]any:
		out := make(map[string]any, len(allowed))
		for k, val := range v {
			if allowed[k] {
				out[k] = val
			}
		}
		return out
	case []any:
		result := make([]any, len(v))
		for i, elem := range v {
			result[i] = FilterFields(elem, fields)
		}
		return result
	}
	return data
}

// SummaryResult is the JSON shape returned by --summary on list commands.
type SummaryResult struct {
	ReturnedItems int  `json:"returned_items"`
	TotalItems    *int `json:"total_items,omitempty"`
	HasMore       bool `json:"has_more"`
}

// JSONSummary writes a SummaryResult as indented JSON to out.
func JSONSummary(out io.Writer, returned int, total *int, hasMore bool) error {
	return JSON(out, SummaryResult{ReturnedItems: returned, TotalItems: total, HasMore: hasMore})
}

// ToGeneric converts any typed value to its generic JSON representation
// (map[string]any, []any, etc.) so that FilterFields can be applied.
func ToGeneric(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out any
	return out, json.Unmarshal(b, &out)
}
