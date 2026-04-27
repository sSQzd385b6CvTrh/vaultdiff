package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func accessLogMetadataResponse(versions map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"versions": versions,
		},
	}
}

func newAccessLogServer(t *testing.T, path string, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(body)
	}))
}

func TestGetAccessLog_Success(t *testing.T) {
	versions := map[string]interface{}{
		"1": map[string]interface{}{"created_time": "2024-01-01T00:00:00Z", "deletion_time": "", "destroyed": false},
		"2": map[string]interface{}{"created_time": "2024-02-01T00:00:00Z", "deletion_time": "", "destroyed": false},
	}
	server := newAccessLogServer(t, "myapp/config", http.StatusOK, accessLogMetadataResponse(versions))
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	logger := NewKVAccessLogger(client, "secret")
	entries, err := logger.GetAccessLog("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Version != 1 {
		t.Errorf("expected first entry version 1, got %d", entries[0].Version)
	}
	if entries[1].Version != 2 {
		t.Errorf("expected second entry version 2, got %d", entries[1].Version)
	}
}

func TestGetAccessLog_NotFound(t *testing.T) {
	server := newAccessLogServer(t, "missing", http.StatusNotFound, map[string]interface{}{})
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	logger := NewKVAccessLogger(client, "secret")
	_, err := logger.GetAccessLog("missing")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewKVAccessLogger_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	logger := NewKVAccessLogger(client, "")
	if logger.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", logger.mount)
	}
}
