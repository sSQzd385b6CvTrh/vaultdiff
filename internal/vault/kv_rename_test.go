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

func newRenameServer(t *testing.T, srcPath, dstPath string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "data/"+srcPath):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "value"},
				},
			})
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "data/"+dstPath):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "metadata/"+srcPath):
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestKVRename_Success(t *testing.T) {
	srv := newRenameServer(t, "old-secret", "new-secret")
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	r := NewKVRenamer(client, "secret")
	err := r.Rename(context.Background(), "old-secret", "new-secret")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestKVRename_SourceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	r := NewKVRenamer(client, "secret")
	err := r.Rename(context.Background(), "missing", "new-secret")
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestNewKVRenamer_DefaultMount(t *testing.T) {
	client, _ := api.NewClient(api.DefaultConfig())
	r := NewKVRenamer(client, "")
	if r.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", r.mount)
	}
}
