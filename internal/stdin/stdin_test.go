package stdin_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/activtrak-mfinlayson/atadmin/internal/stdin"
)

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// errReader always returns an error on Read.
type errReader struct{}

func (e errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestReadJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   io.Reader
		wantErr string
		wantVal testStruct
	}{
		{
			name:    "empty reader",
			input:   bytes.NewReader(nil),
			wantErr: "--from-stdin: stdin is empty; pipe a JSON payload",
		},
		{
			name:    "invalid JSON syntax",
			input:   strings.NewReader("{not json}"),
			wantErr: "--from-stdin: invalid JSON:",
		},
		{
			name:    "valid JSON object",
			input:   strings.NewReader(`{"name":"alice","age":30}`),
			wantVal: testStruct{Name: "alice", Age: 30},
		},
		{
			name:    "read error",
			input:   errReader{},
			wantErr: "--from-stdin: reading stdin:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := stdin.ReadJSON[testStruct](tc.input)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantVal {
				t.Errorf("got %+v, want %+v", got, tc.wantVal)
			}
		})
	}
}

func TestReadRecords(t *testing.T) {
	tests := []struct {
		name    string
		input   io.Reader
		wantErr string
		wantLen int
	}{
		{
			name:    "empty reader",
			input:   bytes.NewReader(nil),
			wantErr: "--from-stdin: stdin is empty; pipe a JSON array",
		},
		{
			name:    "invalid JSON",
			input:   strings.NewReader("not json at all"),
			wantErr: "--from-stdin: invalid JSON:",
		},
		{
			name:    "valid JSON array",
			input:   strings.NewReader(`[{"id":1},{"id":2}]`),
			wantLen: 2,
		},
		{
			name:    "read error",
			input:   errReader{},
			wantErr: "--from-stdin: reading stdin:",
		},
		{
			name:    "JSON object instead of array",
			input:   strings.NewReader(`{"id":1}`),
			wantErr: "--from-stdin: invalid JSON:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := stdin.ReadRecords(tc.input)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tc.wantLen {
				t.Errorf("got %d records, want %d", len(got), tc.wantLen)
			}
		})
	}
}
