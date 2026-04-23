package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func auditTrailMetadataResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"versions": map[string]interface{}{
				"1": map[string]interface{}{
					"created_time":  "2024-01-01T00:00:00Z",
					"deletion_time": "",
					"destroyed":     false,
				},
				"2": map[string]interface{}{
					"created_time":  "2024-02-01T00:00:00Z",
					"deletion_time": "2024-03-01T00:00:00Z",
					"destroyed":     false,
				},
				"3": map[string]interface{}{
					"created_time":  "2024-04-01T00:00:00Z",
					"deletion_time": "",
					"destroyed":     true,
				},
			},
		},
	}
}

func newAuditTrailServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(body)
	}))
}

func TestGetAuditTrail_Success(t *testing.T) {
	srv := newAuditTrailServer(t, http.StatusOK, auditTrailMetadataResponse())
	defer srv.Close()

	trailer := NewKVAuditTrailer(srv.URL, "test-token", "secret")
	entries, err := trailer.GetAuditTrail("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Version != 1 {
		t.Errorf("expected version 1 first, got %d", entries[0].Version)
	}
	if entries[1].DeletedTime == nil {
		t.Error("expected version 2 to have deletion time")
	}
	if !entries[2].Destroyed {
		t.Error("expected version 3 to be destroyed")
	}
}

func TestGetAuditTrail_NotFound(t *testing.T) {
	srv := newAuditTrailServer(t, http.StatusNotFound, map[string]string{})
	defer srv.Close()

	trailer := NewKVAuditTrailer(srv.URL, "test-token", "secret")
	_, err := trailer.GetAuditTrail("missing/path")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewKVAuditTrailer_DefaultMount(t *testing.T) {
	trailer := NewKVAuditTrailer("http://localhost:8200", "token", "")
	if trailer.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", trailer.mount)
	}
}
