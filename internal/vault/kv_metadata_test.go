package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func kvMetadataResponse(currentVersion, maxVersions int) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"created_time":         time.Now().UTC().Format(time.RFC3339),
			"updated_time":         time.Now().UTC().Format(time.RFC3339),
			"current_version":      currentVersion,
			"oldest_version":       1,
			"max_versions":         maxVersions,
			"delete_version_after": "0s",
			"custom_metadata":      map[string]string{"env": "prod"},
		},
	}
}

func TestGetMetadata_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(kvMetadataResponse(5, 10))
	}))
	defer ts.Close()

	reader := NewKVMetadataReader(ts.URL, "test-token", "secret")
	meta, err := reader.GetMetadata("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.CurrentVersion != 5 {
		t.Errorf("expected current_version=5, got %d", meta.CurrentVersion)
	}
	if meta.MaxVersions != 10 {
		t.Errorf("expected max_versions=10, got %d", meta.MaxVersions)
	}
	if meta.CustomMetadata["env"] != "prod" {
		t.Errorf("expected custom_metadata env=prod, got %s", meta.CustomMetadata["env"])
	}
}

func TestGetMetadata_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	reader := NewKVMetadataReader(ts.URL, "test-token", "secret")
	_, err := reader.GetMetadata("missing/path")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
}

func TestGetMetadata_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	reader := NewKVMetadataReader(ts.URL, "test-token", "secret")
	_, err := reader.GetMetadata("some/path")
	if err == nil {
		t.Fatal("expected error for unexpected status, got nil")
	}
}

func TestNewKVMetadataReader_DefaultMount(t *testing.T) {
	reader := NewKVMetadataReader("http://localhost:8200", "token", "")
	if reader.mount != "secret" {
		t.Errorf("expected default mount 'secret', got '%s'", reader.mount)
	}
}
