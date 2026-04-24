package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func freezeMetadataResponse(frozen string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"custom_metadata": map[string]interface{}{
				"frozen": frozen,
			},
		},
	}
}

func newFreezeServer(t *testing.T, statusCode int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestFreeze_Success(t *testing.T) {
	srv := newFreezeServer(t, http.StatusNoContent, nil)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := vaultapi.NewClient(cfg)

	freezer := NewKVFreezer(c, "secret")
	res, err := freezer.Freeze("myapp/config")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Frozen {
		t.Errorf("expected Frozen=true")
	}
	if res.Path != "myapp/config" {
		t.Errorf("unexpected path: %s", res.Path)
	}
}

func TestUnfreeze_Success(t *testing.T) {
	srv := newFreezeServer(t, http.StatusNoContent, nil)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := vaultapi.NewClient(cfg)

	freezer := NewKVFreezer(c, "secret")
	res, err := freezer.Unfreeze("myapp/config")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Frozen {
		t.Errorf("expected Frozen=false")
	}
}

func TestFreeze_NotFound(t *testing.T) {
	srv := newFreezeServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := vaultapi.NewClient(cfg)

	freezer := NewKVFreezer(c, "secret")
	_, err := freezer.Freeze("missing/path")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewKVFreezer_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	c, _ := vaultapi.NewClient(cfg)
	f := NewKVFreezer(c, "")
	if f.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", f.mount)
	}
}
