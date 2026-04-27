package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func bulkGetKVResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
		},
	}
}

func newBulkGetServer(t *testing.T, routes map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := routes[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(body)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestKVBulkGet_AllFound(t *testing.T) {
	routes := map[string]interface{}{
		"/v1/secret/data/alpha": bulkGetKVResponse(map[string]interface{}{"user": "admin"}),
		"/v1/secret/data/beta":  bulkGetKVResponse(map[string]interface{}{"pass": "secret"}),
	}
	srv := newBulkGetServer(t, routes)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "")
	getter := NewKVBulkGetter(client, "secret")
	results := getter.Get(context.Background(), []string{"alpha", "beta"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for key %s: %v", r.Key, r.Err)
		}
	}
}

func TestKVBulkGet_PartialNotFound(t *testing.T) {
	routes := map[string]interface{}{
		"/v1/secret/data/alpha": bulkGetKVResponse(map[string]interface{}{"user": "admin"}),
	}
	srv := newBulkGetServer(t, routes)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "")
	getter := NewKVBulkGetter(client, "secret")
	results := getter.Get(context.Background(), []string{"alpha", "missing"})

	if results[0].Err != nil {
		t.Errorf("expected no error for alpha, got %v", results[0].Err)
	}
	if results[1].Err == nil {
		t.Error("expected error for missing key")
	}
}

func TestNewKVBulkGetter_DefaultMount(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "", "")
	g := NewKVBulkGetter(client, "")
	if g.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", g.mount)
	}
}
