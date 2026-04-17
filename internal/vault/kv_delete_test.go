package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newDeleteServer(t *testing.T, deleteStatus, destroyStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/secret/delete/myapp":
			w.WriteHeader(deleteStatus)
			if deleteStatus == http.StatusOK {
				json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
			}
		case r.Method == http.MethodPut && r.URL.Path == "/v1/secret/destroy/myapp":
			w.WriteHeader(destroyStatus)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestKVDelete_Success(t *testing.T) {
	srv := newDeleteServer(t, http.StatusOK, http.StatusNoContent)
	defer srv.Close()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)
	d := NewKVDeleter(client, "secret")
	if err := d.Delete("myapp", []int{1, 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKVDestroy_Success(t *testing.T) {
	srv := newDeleteServer(t, http.StatusOK, http.StatusNoContent)
	defer srv.Close()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)
	d := NewKVDeleter(client, "secret")
	if err := d.Destroy("myapp", []int{1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKVDestroy_ServerError(t *testing.T) {
	srv := newDeleteServer(t, http.StatusOK, http.StatusInternalServerError)
	defer srv.Close()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)
	d := NewKVDeleter(client, "secret")
	if err := d.Destroy("myapp", []int{1}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewKVDeleter_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	d := NewKVDeleter(client, "")
	if d.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", d.mount)
	}
}
