package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kvCountResponse(keys []string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"keys": keys,
		},
	}
}

func newKVCountServer(t *testing.T, status int, keys []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == 200 {
			json.NewEncoder(w).Encode(kvCountResponse(keys))
		}
	}))
}

func TestKVCount_Success(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	srv := newKVCountServer(t, 200, keys)
	defer srv.Close()

	counter := NewKVCounter(srv.Client(), srv.URL, "test-token", "secret")
	result, err := counter.Count(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 3 {
		t.Errorf("expected count 3, got %d", result.Count)
	}
	if result.Path != "myapp" {
		t.Errorf("expected path 'myapp', got %q", result.Path)
	}
}

func TestKVCount_NotFound(t *testing.T) {
	srv := newKVCountServer(t, 404, nil)
	defer srv.Close()

	counter := NewKVCounter(srv.Client(), srv.URL, "test-token", "secret")
	result, err := counter.Count(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 0 {
		t.Errorf("expected count 0, got %d", result.Count)
	}
}

func TestKVCount_UnexpectedStatus(t *testing.T) {
	srv := newKVCountServer(t, 403, nil)
	defer srv.Close()

	counter := NewKVCounter(srv.Client(), srv.URL, "test-token", "secret")
	_, err := counter.Count(context.Background(), "forbidden")
	if err == nil {
		t.Fatal("expected error for 403, got nil")
	}
}

func TestNewKVCounter_DefaultMount(t *testing.T) {
	counter := NewKVCounter(http.DefaultClient, "http://localhost:8200", "token", "")
	if counter.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", counter.mount)
	}
}
