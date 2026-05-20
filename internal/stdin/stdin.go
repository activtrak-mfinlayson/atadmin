package stdin

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// EnsurePiped returns an error if os.Stdin is an interactive terminal, preventing
// the CLI from hanging silently when --from-stdin is set without piped input.
func EnsurePiped() error {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("--from-stdin: requires data to be piped to the command")
	}
	return nil
}

// ReadJSON reads all bytes from r and unmarshals them into T.
func ReadJSON[T any](r io.Reader) (T, error) {
	var zero T
	data, err := io.ReadAll(r)
	if err != nil {
		return zero, fmt.Errorf("--from-stdin: reading stdin: %w", err)
	}
	if len(data) == 0 {
		return zero, fmt.Errorf("--from-stdin: stdin is empty; pipe a JSON payload")
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, fmt.Errorf("--from-stdin: invalid JSON: %w", err)
	}
	return result, nil
}

// ReadRecords reads all bytes from r and unmarshals them as a JSON array of objects.
// Compatible with the []map[string]any format produced by bulk.ParseFile for JSON files.
func ReadRecords(r io.Reader) ([]map[string]any, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("--from-stdin: reading stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("--from-stdin: stdin is empty; pipe a JSON array")
	}
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("--from-stdin: invalid JSON: %w", err)
	}
	return records, nil
}
