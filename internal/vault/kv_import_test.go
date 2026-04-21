package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newImportServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(statusCode)
	}))
}

func TestKVImport_Success(t *testing.T) {
	srv := newImportServer(t, http.StatusOK)
	defer srv.Close()

	client := &Client{Address: srv.URL, Token: "test-token", HTTPClient: srv.Client()}
	importer := NewKVImporter(client, "secret")

	secrets := map[string]map[string]string{
		"app/db": {"password": "s3cr3t"},
		"app/api": {"key": "abc123"},
	}

	results := importer.Import(context.Background(), secrets)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for path %q, got error: %v", r.Path, r.Error)
		}
	}
}

func TestKVImport_ServerError(t *testing.T) {
	srv := newImportServer(t, http.StatusInternalServerError)
	defer srv.Close()

	client := &Client{Address: srv.URL, Token: "test-token", HTTPClient: srv.Client()}
	importer := NewKVImporter(client, "secret")

	secrets := map[string]map[string]string{
		"app/broken": {"val": "x"},
	}

	results := importer.Import(context.Background(), secrets)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Success {
		t.Error("expected failure, got success")
	}
	if results[0].Error == nil {
		t.Error("expected non-nil error")
	}
}

func TestNewKVImporter_DefaultMount(t *testing.T) {
	client := &Client{}
	importer := NewKVImporter(client, "")
	if importer.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", importer.mount)
	}
}
