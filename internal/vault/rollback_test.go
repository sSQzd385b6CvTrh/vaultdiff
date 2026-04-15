package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

func rollbackKVResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{"version": 3},
		},
	}
}

func newRollbackServer(t *testing.T, readStatus int, writeStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || strings.Contains(r.URL.RawQuery, "version") {
			w.WriteHeader(readStatus)
			if readStatus == http.StatusOK {
				json.NewEncoder(w).Encode(rollbackKVResponse(map[string]interface{}{"key": "val"}))
			}
			return
		}
		w.WriteHeader(writeStatus)
		if writeStatus == http.StatusOK {
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 4}})
		}
	}))
}

func TestRollback_Success(t *testing.T) {
	server := newRollbackServer(t, http.StatusOK, http.StatusOK)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	rb := NewRollbacker(client, "secret")
	res, err := rb.Rollback(context.Background(), "myapp/config", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ToVersion != 3 {
		t.Errorf("expected ToVersion 3, got %d", res.ToVersion)
	}
	if res.Path != "myapp/config" {
		t.Errorf("expected path myapp/config, got %s", res.Path)
	}
}

func TestRollback_NotFound(t *testing.T) {
	server := newRollbackServer(t, http.StatusNotFound, http.StatusOK)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	rb := NewRollbacker(client, "secret")
	_, err := rb.Rollback(context.Background(), "myapp/config", 99)
	if err == nil {
		t.Fatal("expected error for not-found version, got nil")
	}
}

func TestNewRollbacker_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	rb := NewRollbacker(client, "")
	if rb.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", rb.mount)
	}
}
