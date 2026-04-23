package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func snapshotDiffKVResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
		},
	}
}

func newSnapshotDiffServer(pathA, pathB map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp interface{}
		switch r.URL.Path {
		case "/v1/secret/data/a":
			resp = snapshotDiffKVResponse(pathA)
		case "/v1/secret/data/b":
			resp = snapshotDiffKVResponse(pathB)
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestKVSnapshotDiff_Success(t *testing.T) {
	srv := newSnapshotDiffServer(
		map[string]interface{}{"key1": "val1", "key2": "old"},
		map[string]interface{}{"key1": "val1", "key2": "new", "key3": "added"},
	)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	d := NewKVSnapshotDiffer(client, "secret")
	result, err := d.Compare("a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Unchanged["key1"] != "val1" {
		t.Errorf("expected key1 unchanged")
	}
	if result.Modified["key2"] != "new" {
		t.Errorf("expected key2 modified")
	}
	if result.Added["key3"] != "added" {
		t.Errorf("expected key3 added")
	}
}

func TestKVSnapshotDiff_SourceNotFound(t *testing.T) {
	srv := newSnapshotDiffServer(nil, nil)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	d := NewKVSnapshotDiffer(client, "secret")
	_, err := d.Compare("missing", "b")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestNewKVSnapshotDiffer_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	d := NewKVSnapshotDiffer(client, "")
	if d.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", d.mount)
	}
}

func TestKVSnapshotDiff_Removed(t *testing.T) {
	srv := newSnapshotDiffServer(
		map[string]interface{}{"gone": "bye", "stay": "here"},
		map[string]interface{}{"stay": "here"},
	)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	d := NewKVSnapshotDiffer(client, "secret")
	result, err := d.Compare("a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Removed["gone"]; !ok {
		t.Errorf("expected 'gone' in removed")
	}
	if _, ok := result.Unchanged["stay"]; !ok {
		t.Errorf("expected 'stay' in unchanged")
	}
}
