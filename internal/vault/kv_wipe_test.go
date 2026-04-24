package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newWipeServer(t *testing.T, wiped map[string]bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		for path, ok := range wiped {
			if r.URL.Path == "/v1/secret/metadata/"+path {
				if ok {
					w.WriteHeader(http.StatusNoContent)
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{"not found"}})
				}
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestKVWipe_Success(t *testing.T) {
	srv := newWipeServer(t, map[string]bool{"myapp/db": true, "myapp/api": true})
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := api.NewClient(cfg)

	wiper := NewKVWiper(c, "secret")
	results, err := wiper.Wipe([]string{"myapp/db", "myapp/api"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Wiped {
			t.Errorf("expected path %s to be wiped", r.Path)
		}
	}
}

func TestKVWipe_PartialNotFound(t *testing.T) {
	srv := newWipeServer(t, map[string]bool{"myapp/db": true, "myapp/missing": false})
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := api.NewClient(cfg)

	wiper := NewKVWiper(c, "secret")
	results, err := wiper.Wipe([]string{"myapp/db", "myapp/missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestNewKVWiper_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	c, _ := api.NewClient(cfg)
	w := NewKVWiper(c, "")
	if w.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", w.mount)
	}
}
