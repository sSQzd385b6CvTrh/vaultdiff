package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func namespaceListResponse(keys []string) []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"keys": keys,
		},
	})
	return body
}

func TestListNamespaces_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "LIST" {
			t.Errorf("expected LIST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(namespaceListResponse([]string{"team-a/", "team-b/"}))
	}))
	defer ts.Close()

	lister := NewNamespaceLister(ts.URL, "test-token")
	ns, err := lister.ListNamespaces("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ns) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(ns))
	}
	if ns[0].Path != "team-a/" {
		t.Errorf("expected team-a/, got %s", ns[0].Path)
	}
}

func TestListNamespaces_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	lister := NewNamespaceLister(ts.URL, "test-token")
	_, err := lister.ListNamespaces("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListNamespaces_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	lister := NewNamespaceLister(ts.URL, "test-token")
	_, err := lister.ListNamespaces("")
	if err == nil {
		t.Fatal("expected error for forbidden, got nil")
	}
}

func TestListNamespaces_WithParent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/engineering/sys/namespaces" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(namespaceListResponse([]string{"backend/"}))
	}))
	defer ts.Close()

	lister := NewNamespaceLister(ts.URL, "test-token")
	ns, err := lister.ListNamespaces("engineering")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ns) != 1 || ns[0].Path != "backend/" {
		t.Errorf("unexpected namespaces: %+v", ns)
	}
}
