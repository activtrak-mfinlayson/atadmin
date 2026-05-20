package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
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
// Uses reflection to avoid double-serialization overhead.
func ToGeneric(v any) (any, error) {
	return reflectToGeneric(reflect.ValueOf(v))
}

func reflectToGeneric(rv reflect.Value) (any, error) {
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Struct:
		rt := rv.Type()
		m := make(map[string]any, rt.NumField())
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			fv := rv.Field(i)
			if !field.IsExported() {
				continue
			}
			name, omitempty := jsonFieldInfo(field)
			if name == "-" {
				continue
			}
			if omitempty && isEmptyReflectValue(fv) {
				continue
			}
			val, err := reflectToGeneric(fv)
			if err != nil {
				return nil, err
			}
			m[name] = val
		}
		return m, nil
	case reflect.Slice:
		if rv.IsNil() {
			return nil, nil
		}
		result := make([]any, rv.Len())
		for i := range result {
			val, err := reflectToGeneric(rv.Index(i))
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	case reflect.Map:
		if rv.IsNil() {
			return nil, nil
		}
		m := make(map[string]any, rv.Len())
		for _, key := range rv.MapKeys() {
			val, err := reflectToGeneric(rv.MapIndex(key))
			if err != nil {
				return nil, err
			}
			m[fmt.Sprintf("%v", key.Interface())] = val
		}
		return m, nil
	default:
		return rv.Interface(), nil
	}
}

func jsonFieldInfo(f reflect.StructField) (name string, omitempty bool) {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name, false
	}
	parts := strings.SplitN(tag, ",", 2)
	if parts[0] == "" {
		name = f.Name
	} else {
		name = parts[0]
	}
	if len(parts) > 1 && strings.Contains(parts[1], "omitempty") {
		omitempty = true
	}
	return
}

func isEmptyReflectValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}
