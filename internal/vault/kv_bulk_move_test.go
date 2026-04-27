package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func bulkMoveKVResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
		},
	}
}

func newBulkMoveServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]map[string]interface{}{
		"src/alpha": {"key": "val1"},
		"src/beta":  {"key": "val2"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			for path, d := range store {
				if r.URL.Path == "/v1/secret/data/"+path {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(bulkMoveKVResponse(d))
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPost, http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestKVBulkMove_AllSuccess(t *testing.T) {
	srv := newBulkMoveServer(t)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "", "")
	mover := NewKVBulkMover(client, "secret")
	pairs := []MovePair{
		{Source: "src/alpha", Dest: "dst/alpha"},
		{Source: "src/beta", Dest: "dst/beta"},
	}
	results := mover.Move(pairs)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s→%s: %v", r.Source, r.Dest, r.Err)
		}
	}
}

func TestKVBulkMove_SourceNotFound(t *testing.T) {
	srv := newBulkMoveServer(t)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "", "")
	mover := NewKVBulkMover(client, "secret")
	results := mover.Move([]MovePair{{Source: "missing/key", Dest: "dst/key"}})
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if results[0].Err == nil {
		t.Error("expected error for missing source, got nil")
	}
}

func TestKVBulkMove_SameSourceAndDest(t *testing.T) {
	srv := newBulkMoveServer(t)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "", "")
	mover := NewKVBulkMover(client, "secret")
	results := mover.Move([]MovePair{{Source: "src/alpha", Dest: "src/alpha"}})
	if results[0].Err == nil {
		t.Error("expected error when source == dest")
	}
}

func TestNewKVBulkMover_DefaultMount(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "", "")
	m := NewKVBulkMover(client, "")
	if m.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", m.mount)
	}
}
