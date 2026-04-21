package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func cloneKVResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{"version": json.Number("1")},
		},
	}
}

func newCloneServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/src/key":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cloneKVResponse(map[string]interface{}{"foo": "bar"}))
		case "/v1/secret/data/dst/key":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 2}})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestKVClone_Success(t *testing.T) {
	srv := newCloneServer(t)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	cloner := NewKVCloner(client, "secret")
	result, err := cloner.Clone("src/key", "dst/key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestKVClone_SourceNotFound(t *testing.T) {
	srv := newCloneServer(t)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	cloner := NewKVCloner(client, "secret")
	_, err := cloner.Clone("missing/key", "dst/key", "")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestNewKVCloner_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	cloner := NewKVCloner(client, "")
	if cloner.mount != "secret" {
		t.Errorf("expected mount=secret, got %q", cloner.mount)
	}
}
