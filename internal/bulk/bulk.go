package bulk

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParseFile reads a JSON or CSV file and returns its records as a slice of maps.
// Format is auto-detected by file extension (.json or .csv).
func ParseFile(path string) ([]map[string]any, error) {
	ext := strings.ToLower(filepath.Ext(path))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", path, err)
	}

	switch ext {
	case ".json":
		return parseJSON(data)
	case ".csv":
		return parseCSV(data)
	default:
		return nil, fmt.Errorf("unsupported file format %q: expected .json or .csv", ext)
	}
}

func parseJSON(data []byte) ([]map[string]any, error) {
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return records, nil
}

func parseCSV(data []byte) ([]map[string]any, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV headers: %w", err)
	}

	var records []map[string]any
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		m := make(map[string]any, len(headers))
		for i, h := range headers {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		records = append(records, m)
	}
	return records, nil
}
