package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kvPutResponse(version int) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{"version": version},
	}
}

func TestKVPut_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(kvPutResponse(3))
	}))
	defer ts.Close()

	w := NewKVWriter(ts.URL, "token", "secret")
	ver, err := w.Put("myapp/config", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver != 3 {
		t.Errorf("expected version 3, got %d", ver)
	}
}

func TestKVPut_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer ts.Close()

	w := NewKVWriter(ts.URL, "bad-token", "secret")
	_, err := w.Put("myapp/config", map[string]string{"key": "val"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewKVWriter_DefaultMount(t *testing.T) {
	w := NewKVWriter("http://localhost:8200", "token", "")
	if w.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", w.mount)
	}
}
