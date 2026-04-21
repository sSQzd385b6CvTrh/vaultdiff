package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newRestoreVersionServer(t *testing.T, path string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
}

func TestRestoreVersion_Success(t *testing.T) {
	srv := newRestoreVersionServer(t, "myapp/config", http.StatusNoContent)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)

	r := NewKVVersionRestorer(client, "secret")
	err := r.RestoreVersion(context.Background(), "myapp/config", []int{2, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRestoreVersion_NotFound(t *testing.T) {
	srv := newRestoreVersionServer(t, "missing/path", http.StatusNotFound)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)

	r := NewKVVersionRestorer(client, "secret")
	err := r.RestoreVersion(context.Background(), "missing/path", []int{1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRestoreVersion_NoVersions(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	client, _ := vaultapi.NewClient(cfg)

	r := NewKVVersionRestorer(client, "secret")
	err := r.RestoreVersion(context.Background(), "myapp/config", []int{})
	if err == nil {
		t.Fatal("expected error for empty versions")
	}
}

func TestNewKVVersionRestorer_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	client, _ := vaultapi.NewClient(cfg)

	r := NewKVVersionRestorer(client, "")
	if r.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", r.mount)
	}
}
