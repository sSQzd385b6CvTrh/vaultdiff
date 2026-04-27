package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func bulkCopyKVResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{"version": 1},
		},
	}
}

func newBulkCopyServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]map[string]interface{}{
		"alpha": {"key": "value1"},
		"beta":  {"key": "value2"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			path := r.URL.Path[len("/v1/secret/data/"):]
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(bulkCopyKVResponse(data))
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 1}})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestKVBulkCopy_AllSuccess(t *testing.T) {
	srv := newBulkCopyServer(t)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "")
	copier := NewKVBulkCopier(client, "secret")

	pairs := []CopyPair{
		{Source: "alpha", Destination: "alpha-copy"},
		{Source: "beta", Destination: "beta-copy"},
	}
	results := copier.Copy(context.Background(), pairs)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s→%s: %v", r.Source, r.Destination, r.Err)
		}
	}
}

func TestKVBulkCopy_PartialNotFound(t *testing.T) {
	srv := newBulkCopyServer(t)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "")
	copier := NewKVBulkCopier(client, "secret")

	pairs := []CopyPair{
		{Source: "alpha", Destination: "alpha-copy"},
		{Source: "missing", Destination: "missing-copy"},
	}
	results := copier.Copy(context.Background(), pairs)
	if results[0].Err != nil {
		t.Errorf("expected success for alpha, got %v", results[0].Err)
	}
	if results[1].Err == nil {
		t.Error("expected error for missing source, got nil")
	}
}

func TestNewKVBulkCopier_DefaultMount(t *testing.T) {
	client, _ := NewClient("http://localhost:8200", "", "")
	copier := NewKVBulkCopier(client, "")
	if copier.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", copier.mount)
	}
}
