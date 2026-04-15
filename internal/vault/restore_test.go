package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func restoreKVResponse(version int) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"version": version,
		},
	}
}

func newRestoreServer(t *testing.T, statusCode int, version int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			json.NewEncoder(w).Encode(restoreKVResponse(version))
		}
	}))
}

func TestRestoreSnapshot_Success(t *testing.T) {
	server := newRestoreServer(t, http.StatusOK, 3)
	defer server.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	restorer := NewRestorer(client, "secret")
	snapshot := map[string]map[string]string{
		"app/config": {"DB_HOST": "localhost", "DB_PORT": "5432"},
	}

	results := restorer.RestoreSnapshot(context.Background(), snapshot)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}
	if results[0].Key != "app/config" {
		t.Errorf("expected key app/config, got %s", results[0].Key)
	}
}

func TestRestoreSnapshot_ServerError(t *testing.T) {
	server := newRestoreServer(t, http.StatusInternalServerError, 0)
	defer server.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	restorer := NewRestorer(client, "secret")
	snapshot := map[string]map[string]string{
		"app/config": {"KEY": "value"},
	}

	results := restorer.RestoreSnapshot(context.Background(), snapshot)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestRestoreSnapshot_Empty(t *testing.T) {
	restorer := NewRestorer(nil, "secret")
	results := restorer.RestoreSnapshot(context.Background(), map[string]map[string]string{})
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty snapshot, got %d", len(results))
	}
}
