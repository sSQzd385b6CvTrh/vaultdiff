package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func protectMetadataResponse(protected string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"custom_metadata": map[string]interface{}{"protected": protected},
		},
	}
}

func newProtectServer(t *testing.T, getStatus int, getBody interface{}, postStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(getStatus)
			if getBody != nil {
				json.NewEncoder(w).Encode(getBody)
			}
		case http.MethodPost:
			w.WriteHeader(postStatus)
		}
	}))
}

func TestProtect_Success(t *testing.T) {
	srv := newProtectServer(t, http.StatusOK, nil, http.StatusNoContent)
	defer srv.Close()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := api.NewClient(cfg)
	p := NewKVProtector(c, "secret")
	if err := p.Protect("myapp/config"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnprotect_Success(t *testing.T) {
	srv := newProtectServer(t, http.StatusOK, nil, http.StatusNoContent)
	defer srv.Close()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := api.NewClient(cfg)
	p := NewKVProtector(c, "secret")
	if err := p.Unprotect("myapp/config"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsProtected_True(t *testing.T) {
	body := protectMetadataResponse("true")
	srv := newProtectServer(t, http.StatusOK, body, http.StatusNoContent)
	defer srv.Close()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := api.NewClient(cfg)
	p := NewKVProtector(c, "secret")
	ok, err := p.IsProtected("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected protected=true")
	}
}

func TestIsProtected_NotFound(t *testing.T) {
	srv := newProtectServer(t, http.StatusNotFound, nil, http.StatusNoContent)
	defer srv.Close()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := api.NewClient(cfg)
	p := NewKVProtector(c, "secret")
	ok, err := p.IsProtected("missing/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected protected=false for missing path")
	}
}

func TestNewKVProtector_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	c, _ := api.NewClient(cfg)
	p := NewKVProtector(c, "")
	if p.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", p.mount)
	}
}
