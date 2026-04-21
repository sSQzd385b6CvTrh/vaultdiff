package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func archiveMetaResponse(versions map[string]interface{}) []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"versions": versions,
		},
	})
	return body
}

func newArchiveServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/myapp/db":
			w.Header().Set("Content-Type", "application/json")
			w.Write(archiveMetaResponse(map[string]interface{}{
				"1": map[string]interface{}{"destroyed": false, "deletion_time": ""},
				"2": map[string]interface{}{"destroyed": true, "deletion_time": ""},
			}))
		case "/v1/secret/data/myapp/db":
			w.Header().Set("Content-Type", "application/json")
			body, _ := json.Marshal(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"user": "admin"},
				},
			})
			w.Write(body)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestKVArchive_Success(t *testing.T) {
	srv := newArchiveServer(t)
	defer srv.Close()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)
	a := NewKVArchiver(client, "secret")
	entry, err := a.Archive(context.Background(), "myapp/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Key != "myapp/db" {
		t.Errorf("expected key myapp/db, got %s", entry.Key)
	}
	if _, ok := entry.Versions["1"]; !ok {
		t.Error("expected version 1 to be present")
	}
	if _, ok := entry.Versions["2"]; ok {
		t.Error("destroyed version 2 should be skipped")
	}
}

func TestKVArchive_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)
	a := NewKVArchiver(client, "secret")
	_, err := a.Archive(context.Background(), "missing/key")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestNewKVArchiver_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	client, _ := vaultapi.NewClient(cfg)
	a := NewKVArchiver(client, "")
	if a.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", a.mount)
	}
}
