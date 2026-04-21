package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newBulkDeleteServer(t *testing.T, statuses map[string]int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		for path, code := range statuses {
			if r.URL.Path == "/v1/secret/data/"+path {
				w.WriteHeader(code)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestKVBulkDelete_AllSuccess(t *testing.T) {
	statuses := map[string]int{
		"apps/alpha": http.StatusNoContent,
		"apps/beta":  http.StatusNoContent,
	}
	srv := newBulkDeleteServer(t, statuses)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := vaultapi.NewClient(cfg)

	deleter := NewKVBulkDeleter(c, "secret")
	results := deleter.DeleteAll(context.Background(), []string{"apps/alpha", "apps/beta"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Deleted {
			t.Errorf("expected %q to be deleted", r.Path)
		}
		if r.Err != nil {
			t.Errorf("unexpected error for %q: %v", r.Path, r.Err)
		}
	}
}

func TestKVBulkDelete_PartialNotFound(t *testing.T) {
	statuses := map[string]int{
		"apps/alpha": http.StatusNoContent,
	}
	srv := newBulkDeleteServer(t, statuses)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := vaultapi.NewClient(cfg)

	deleter := NewKVBulkDeleter(c, "secret")
	results := deleter.DeleteAll(context.Background(), []string{"apps/alpha", "apps/missing"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Deleted {
		t.Errorf("expected apps/alpha to be deleted")
	}
	if results[1].Err == nil {
		t.Errorf("expected error for apps/missing")
	}
}

func TestNewKVBulkDeleter_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	c, _ := vaultapi.NewClient(cfg)
	d := NewKVBulkDeleter(c, "")
	if d.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", d.mount)
	}
}
