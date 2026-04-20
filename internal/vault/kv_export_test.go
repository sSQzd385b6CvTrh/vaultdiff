package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newExportServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/secret/metadata/myapp" && r.URL.Query().Get("list") == "true":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"db", "api"}},
			})
		case r.URL.Path == "/v1/secret/data/myapp/db":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"password": "s3cr3t"}},
			})
		case r.URL.Path == "/v1/secret/data/myapp/api":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"token": "abc123"}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestKVExport_Success(t *testing.T) {
	srv := newExportServer(t)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	exporter := NewKVExporter(client, "secret")
	result, err := exporter.Export("myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(result))
	}
	if result["myapp/db"]["password"] != "s3cr3t" {
		t.Errorf("expected password s3cr3t, got %v", result["myapp/db"]["password"])
	}
}

func TestKVExport_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	exporter := NewKVExporter(client, "secret")
	_, err := exporter.Export("missing")
	if err == nil {
		t.Fatal("expected error for not found path")
	}
}

func TestNewKVExporter_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	e := NewKVExporter(client, "")
	if e.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", e.mount)
	}
}
