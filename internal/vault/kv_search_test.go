package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

func kvSearchListResponse(keys []string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{"keys": keys},
	}
}

func kvSearchDataResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{"data": data},
	}
}

func newSearchServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/metadata") {
			json.NewEncoder(w).Encode(kvSearchListResponse([]string{"db", "api"}))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/data/app/db") {
			json.NewEncoder(w).Encode(kvSearchDataResponse(map[string]interface{}{
				"db_password": "secret",
				"host":        "localhost",
			}))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/data/app/api") {
			json.NewEncoder(w).Encode(kvSearchDataResponse(map[string]interface{}{
				"api_key": "abc",
				"timeout": "30",
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestKVSearch_MatchesKey(t *testing.T) {
	srv := newSearchServer(t)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	searcher := NewKVSearcher(client, "secret")
	results, err := searcher.Search(context.Background(), "app", "db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "app/db" {
		t.Errorf("expected path app/db, got %s", results[0].Path)
	}
}

func TestKVSearch_NoMatch(t *testing.T) {
	srv := newSearchServer(t)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	searcher := NewKVSearcher(client, "secret")
	results, err := searcher.Search(context.Background(), "app", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestNewKVSearcher_DefaultMount(t *testing.T) {
	client, _ := api.NewClient(api.DefaultConfig())
	s := NewKVSearcher(client, "")
	if s.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", s.mount)
	}
}
