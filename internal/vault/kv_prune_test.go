package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func pruneMetadataResponse(versions map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"versions": versions,
		},
	}
}

func newPruneServer(t *testing.T, versions map[string]interface{}, destroyStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pruneMetadataResponse(versions))
		case http.MethodPut:
			w.WriteHeader(destroyStatus)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestKVPrune_Success(t *testing.T) {
	versions := map[string]interface{}{
		"1": map[string]interface{}{"destroyed": false},
		"2": map[string]interface{}{"destroyed": false},
		"3": map[string]interface{}{"destroyed": false},
		"4": map[string]interface{}{"destroyed": false},
	}
	srv := newPruneServer(t, versions, http.StatusNoContent)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "")
	pruner := NewKVPruner(client, "secret")

	result, err := pruner.Prune(context.Background(), "myapp/config", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pruned != 2 {
		t.Errorf("expected 2 pruned, got %d", result.Pruned)
	}
	if result.Skipped != 2 {
		t.Errorf("expected 2 skipped, got %d", result.Skipped)
	}
}

func TestKVPrune_NothingToPrune(t *testing.T) {
	versions := map[string]interface{}{
		"1": map[string]interface{}{"destroyed": false},
	}
	srv := newPruneServer(t, versions, http.StatusNoContent)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "")
	pruner := NewKVPruner(client, "secret")

	result, err := pruner.Prune(context.Background(), "myapp/config", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pruned != 0 {
		t.Errorf("expected 0 pruned, got %d", result.Pruned)
	}
	if result.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Skipped)
	}
}

func TestKVPrune_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, "", "")
	pruner := NewKVPruner(client, "secret")

	_, err := pruner.Prune(context.Background(), "missing/key", 2)
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestNewKVPruner_DefaultMount(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "", "")
	p := NewKVPruner(client, "")
	if p.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", p.mount)
	}
}
