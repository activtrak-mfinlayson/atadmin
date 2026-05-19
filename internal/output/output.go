package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
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
