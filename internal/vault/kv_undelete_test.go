package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newUndeleteServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func TestKVUndelete_Success(t *testing.T) {
	ts := newUndeleteServer(http.StatusNoContent)
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	u := NewKVUndeleter(client, "secret")
	err := u.Undelete(context.Background(), "myapp/config", []int{1, 2})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestKVUndelete_NotFound(t *testing.T) {
	ts := newUndeleteServer(http.StatusNotFound)
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	u := NewKVUndeleter(client, "secret")
	err := u.Undelete(context.Background(), "missing/path", []int{1})
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
}

func TestKVUndelete_NoVersions(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	client, _ := vaultapi.NewClient(cfg)

	u := NewKVUndeleter(client, "secret")
	err := u.Undelete(context.Background(), "myapp/config", []int{})
	if err == nil {
		t.Fatal("expected error for empty versions, got nil")
	}
}

func TestKVUndelete_Forbidden(t *testing.T) {
	ts := newUndeleteServer(http.StatusForbidden)
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	u := NewKVUndeleter(client, "secret")
	err := u.Undelete(context.Background(), "restricted/path", []int{3})
	if err == nil {
		t.Fatal("expected permission denied error, got nil")
	}
}

func TestNewKVUndeleter_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	client, _ := vaultapi.NewClient(cfg)

	u := NewKVUndeleter(client, "")
	if u.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", u.mount)
	}
}
