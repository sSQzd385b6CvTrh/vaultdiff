package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kvGetResponse(data map[string]interface{}, version int) interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{"version": version},
		},
	}
}

func TestKVGet_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(kvGetResponse(map[string]interface{}{"foo": "bar"}, 3))
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "tok"}
	g := NewKVGetter(c, "secret")
	entry, err := g.Get("myapp/config", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Data["foo"] != "bar" {
		t.Errorf("expected bar, got %s", entry.Data["foo"])
	}
	if entry.Version != 3 {
		t.Errorf("expected version 3, got %d", entry.Version)
	}
}

func TestKVGet_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "tok"}
	g := NewKVGetter(c, "secret")
	_, err := g.Get("missing/path", 0)
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestKVGet_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "tok"}
	g := NewKVGetter(c, "")
	_, err := g.Get("some/path", 0)
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestNewKVGetter_DefaultMount(t *testing.T) {
	c := &Client{Address: "http://localhost:8200", Token: "tok"}
	g := NewKVGetter(c, "")
	if g.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", g.mount)
	}
}
