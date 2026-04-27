package vault

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

func newSetTTLServer(t *testing.T, path string, wantTTL string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		if got, ok := payload["delete_version_after"]; ok {
			if got.(string) != wantTTL {
				t.Errorf("expected TTL %q, got %q", wantTTL, got)
			}
		}
		w.WriteHeader(statusCode)
	}))
}

func TestKVSetTTL_Success(t *testing.T) {
	svr := newSetTTLServer(t, "myapp/config", "24h0m0s", http.StatusNoContent)
	defer svr.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = svr.URL
	client, _ := vaultapi.NewClient(cfg)

	setter := NewKVTTLSetter(client, "secret")
	err := setter.SetTTL(context.Background(), "myapp/config", 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKVSetTTL_Clear(t *testing.T) {
	svr := newSetTTLServer(t, "myapp/config", "", http.StatusNoContent)
	defer svr.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = svr.URL
	client, _ := vaultapi.NewClient(cfg)

	setter := NewKVTTLSetter(client, "secret")
	err := setter.SetTTL(context.Background(), "myapp/config", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKVSetTTL_NotFound(t *testing.T) {
	svr := newSetTTLServer(t, "missing/path", "", http.StatusNotFound)
	defer svr.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = svr.URL
	client, _ := vaultapi.NewClient(cfg)

	setter := NewKVTTLSetter(client, "secret")
	err := setter.SetTTL(context.Background(), "missing/path", time.Hour)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewKVTTLSetter_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	client, _ := vaultapi.NewClient(cfg)
	setter := NewKVTTLSetter(client, "")
	if setter.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", setter.mount)
	}
}
