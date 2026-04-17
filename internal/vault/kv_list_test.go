package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kvListResponse(keys []string) []byte {
	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{"keys": keys},
	})
	return body
}

func TestListKeys_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("list") != "true" {
			t.Error("expected list=true query param")
		}
		w.WriteHeader(http.StatusOK)
		w.Write(kvListResponse([]string{"foo", "bar", "baz"}))
	}))
	defer ts.Close()

	lister := NewKVLister(ts.URL, "test-token", "secret")
	keys, err := lister.ListKeys("myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "foo" {
		t.Errorf("expected foo, got %s", keys[0])
	}
}

func TestListKeys_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	lister := NewKVLister(ts.URL, "test-token", "")
	_, err := lister.ListKeys("missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestListKeys_MissingKeysField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{}}`))
	}))
	defer ts.Close()

	lister := NewKVLister(ts.URL, "test-token", "secret")
	_, err := lister.ListKeys("myapp")
	if err == nil {
		t.Fatal("expected error for missing keys")
	}
}

func TestNewKVLister_DefaultMount(t *testing.T) {
	lister := NewKVLister("http://localhost:8200", "token", "")
	if lister.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", lister.mount)
	}
}
