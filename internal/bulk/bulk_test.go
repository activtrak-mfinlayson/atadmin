package bulk_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/bulk"
)

func writeTemp(t *testing.T, ext, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "bulk*"+ext)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return f.Name()
}

func TestParseFileJSON(t *testing.T) {
	path := writeTemp(t, ".json", `[{"username":"alice","id":"1"}]`)
	records, err := bulk.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0]["username"] != "alice" {
		t.Errorf("expected username=alice, got %v", records[0]["username"])
	}
}

func TestParseFileCSV(t *testing.T) {
	path := writeTemp(t, ".csv", "username,id\nalice,1\nbob,2\n")
	records, err := bulk.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0]["username"] != "alice" {
		t.Errorf("expected username=alice, got %v", records[0]["username"])
	}
}

func TestParseFileUnsupportedExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.xml")
	_ = os.WriteFile(path, []byte("<root/>"), 0600)
	_, err := bulk.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for .xml extension, got nil")
	}
}

func TestParseFileMissing(t *testing.T) {
	_, err := bulk.ParseFile("/nonexistent/path/file.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
