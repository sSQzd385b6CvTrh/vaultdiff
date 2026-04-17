package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newPatchServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(status)
	}))
}

func TestKVPatch_Success(t *testing.T) {
	srv := newPatchServer(t, http.StatusOK)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)

	p := NewKVPatcher(client, "secret")
	err := p.Patch(context.Background(), "myapp/config", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestKVPatch_NotFound(t *testing.T) {
	srv := newPatchServer(t, http.StatusNotFound)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)

	p := NewKVPatcher(client, "secret")
	err := p.Patch(context.Background(), "missing/path", map[string]interface{}{"k": "v"})
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestKVPatch_ServerError(t *testing.T) {
	srv := newPatchServer(t, http.StatusInternalServerError)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)

	p := NewKVPatcher(client, "secret")
	err := p.Patch(context.Background(), "myapp/config", map[string]interface{}{"k": "v"})
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestNewKVPatcher_DefaultMount(t *testing.T) {
	client, _ := vaultapi.NewClient(vaultapi.DefaultConfig())
	p := NewKVPatcher(client, "")
	if p.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", p.mount)
	}
}
