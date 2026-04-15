package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *vaultapi.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vaultapi.NewClient: %v", err)
	}
	return srv, c
}

func kvResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{"version": 1},
		},
	}
}

func TestGetSecretVersion_Success(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(kvResponse(map[string]interface{}{
			"username": "admin",
			"password": "s3cr3t",
		}))
	})

	f := NewFetcher(c)
	sv, err := f.GetSecretVersion(context.Background(), "secret", "myapp/config", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.Data["username"] != "admin" {
		t.Errorf("expected username=admin, got %q", sv.Data["username"])
	}
	if sv.Data["password"] != "s3cr3t" {
		t.Errorf("expected password=s3cr3t, got %q", sv.Data["password"])
	}
}

func TestGetSecretVersion_NotFound(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{}`))
	})

	f := NewFetcher(c)
	_, err := f.GetSecretVersion(context.Background(), "secret", "missing/path", 1)
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}

func TestGetSecretVersion_MissingDataKey(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
	})

	f := NewFetcher(c)
	_, err := f.GetSecretVersion(context.Background(), "secret", "myapp/config", 0)
	if err == nil {
		t.Fatal("expected error for missing 'data' key, got nil")
	}
}
