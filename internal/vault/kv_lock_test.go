package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newLockServer(t *testing.T, statusCode int, capturedBody *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capturedBody != nil {
			if err := json.NewDecoder(r.Body).Decode(capturedBody); err != nil {
				t.Errorf("failed to decode request body: %v", err)
			}
		}
		w.WriteHeader(statusCode)
	}))
}

func TestLock_Success(t *testing.T) {
	var body map[string]interface{}
	srv := newLockServer(t, http.StatusNoContent, &body)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token", "")
	locker := NewKVLocker(client, "secret")

	err := locker.Lock("myapp/config", "alice", 10*time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	cm, ok := body["custom_metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("expected custom_metadata in request body")
	}
	if cm["locked"] != "true" {
		t.Errorf("expected locked=true, got %v", cm["locked"])
	}
	if cm["locked_by"] != "alice" {
		t.Errorf("expected locked_by=alice, got %v", cm["locked_by"])
	}
}

func TestUnlock_Success(t *testing.T) {
	var body map[string]interface{}
	srv := newLockServer(t, http.StatusNoContent, &body)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token", "")
	locker := NewKVLocker(client, "secret")

	err := locker.Unlock("myapp/config")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	cm, ok := body["custom_metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("expected custom_metadata in request body")
	}
	if cm["locked"] != "false" {
		t.Errorf("expected locked=false, got %v", cm["locked"])
	}
}

func TestLock_ServerError(t *testing.T) {
	srv := newLockServer(t, http.StatusForbidden, nil)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "bad-token", "")
	locker := NewKVLocker(client, "")

	err := locker.Lock("myapp/config", "bob", 5*time.Minute)
	if err == nil {
		t.Fatal("expected error for forbidden response")
	}
}

func TestNewKVLocker_DefaultMount(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "token", "")
	locker := NewKVLocker(client, "")
	if locker.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", locker.mount)
	}
}
