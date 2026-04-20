package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMoveServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// GET metadata (allVersions)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/src/key":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"versions": map[string]interface{}{
						"1": map[string]interface{}{},
						"2": map[string]interface{}{},
					},
				},
			})
		// GET data (copy source read)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/src/key":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"foo": "bar"},
				},
			})
		// POST data (copy dest write)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/secret/data/dst/key":
			w.WriteHeader(http.StatusOK)
		// POST destroy (delete all versions)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/secret/destroy/src/key":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestKVMove_Success(t *testing.T) {
	srv := newMoveServer(t)
	defer srv.Close()

	client := &Client{Address: srv.URL, Token: "test-token"}
	mover := NewKVMover(client, "secret")

	if err := mover.Move(context.Background(), "src/key", "dst/key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKVMove_CopyFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := &Client{Address: srv.URL, Token: "test-token"}
	mover := NewKVMover(client, "secret")

	err := mover.Move(context.Background(), "src/key", "dst/key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewKVMover_DefaultMount(t *testing.T) {
	client := &Client{Address: "http://localhost:8200", Token: "tok"}
	m := NewKVMover(client, "")
	if m.mount != "secret" {
		t.Errorf("expected mount 'secret', got %q", m.mount)
	}
}
