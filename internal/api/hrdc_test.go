package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// HRDCPing
// ---------------------------------------------------------------------------

func TestHRDCPing(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	if err := client.HRDCPing(context.Background()); err != nil {
		t.Fatalf("HRDCPing() error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("method = %s, want GET", gotMethod)
	}
	if gotPath != "/hrdc/ping" {
		t.Errorf("path = %s, want /hrdc/ping", gotPath)
	}
}

func TestHRDCPing_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	if err := client.HRDCPing(context.Background()); err == nil {
		t.Fatal("HRDCPing() expected error on 503, got nil")
	}
}

// ---------------------------------------------------------------------------
// HRDCBulkImport
// ---------------------------------------------------------------------------

func TestHRDCBulkImport(t *testing.T) {
	records := []map[string]any{
		{"employeeId": "E001", "name": "Alice"},
		{"employeeId": "E002", "name": "Bob"},
	}

	var gotMethod, gotPath string
	var gotBody struct {
		Records []map[string]any `json:"records"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	if err := client.HRDCBulkImport(context.Background(), records); err != nil {
		t.Fatalf("HRDCBulkImport() error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotPath != "/hrdc/v1/bulk" {
		t.Errorf("path = %s, want /hrdc/v1/bulk", gotPath)
	}

	// Verify the body was wrapped under "records" key.
	if len(gotBody.Records) != len(records) {
		t.Fatalf("body.records len = %d, want %d", len(gotBody.Records), len(records))
	}
}

func TestHRDCBulkImport_BodyStructure(t *testing.T) {
	// Verify that the request body uses {"records": [...]} wrapper structure.
	records := []map[string]any{
		{"key": "value"},
	}

	var rawBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, "test-token")
	_ = client.HRDCBulkImport(context.Background(), records)

	// Top-level key must be "records".
	if _, ok := rawBody["records"]; !ok {
		t.Errorf("request body missing top-level 'records' key, got keys: %v", rawBody)
	}
}
