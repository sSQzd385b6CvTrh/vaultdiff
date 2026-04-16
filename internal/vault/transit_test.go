package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func transitListResponse(keys []string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"keys": keys,
		},
	}
}

func TestListTransitKeys_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/transit/keys" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transitListResponse([]string{"my-key", "another-key"}))
	}))
	defer ts.Close()

	lister := NewTransitLister(ts.URL, "test-token", "")
	keys, err := lister.ListKeys()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "my-key" {
		t.Errorf("expected my-key, got %s", keys[0])
	}
}

func TestListTransitKeys_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	lister := NewTransitLister(ts.URL, "test-token", "transit")
	keys, err := lister.ListKeys()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty list, got %v", keys)
	}
}

func TestListTransitKeys_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	lister := NewTransitLister(ts.URL, "bad-token", "transit")
	_, err := lister.ListKeys()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewTransitLister_DefaultMount(t *testing.T) {
	lister := NewTransitLister("http://localhost:8200", "token", "")
	if lister.Mount != "transit" {
		t.Errorf("expected default mount 'transit', got %s", lister.Mount)
	}
}
