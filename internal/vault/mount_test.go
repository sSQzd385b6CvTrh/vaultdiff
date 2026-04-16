package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mountListResponse(t *testing.T, w http.ResponseWriter, payload interface{}, status int) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func TestListMounts_Success(t *testing.T) {
	payload := map[string]interface{}{
		"secret/": map[string]interface{}{"type": "kv", "description": "KV secrets", "local": false},
		"pki/":    map[string]interface{}{"type": "pki", "description": "PKI engine", "local": true},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/mounts" {
			http.NotFound(w, r)
			return
		}
		mountListResponse(t, w, payload, http.StatusOK)
	}))
	defer ts.Close()

	lister := NewMountLister(ts.URL, "test-token")
	mounts, err := lister.ListMounts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(mounts))
	}
	// sorted: pki/ before secret/
	if mounts[0].Path != "pki/" {
		t.Errorf("expected pki/ first, got %s", mounts[0].Path)
	}
	if !mounts[0].Local {
		t.Errorf("expected pki/ to be local")
	}
	if mounts[1].Type != "kv" {
		t.Errorf("expected secret/ type kv, got %s", mounts[1].Type)
	}
}

func TestListMounts_Forbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	lister := NewMountLister(ts.URL, "bad-token")
	_, err := lister.ListMounts()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListMounts_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	lister := NewMountLister(ts.URL, "token")
	_, err := lister.ListMounts()
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}
