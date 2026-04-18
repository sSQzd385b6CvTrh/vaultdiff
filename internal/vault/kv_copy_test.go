package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCopyServer(t *testing.T, srcPath string, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": data,
				},
			})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 1}})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestKVCopy_Success(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	server := newCopyServer(t, "src/secret", data)
	defer server.Close()

	copier := NewKVCopier(server.URL, "test-token", "secret")
	if err := copier.Copy("src/secret", "dst/secret"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestKVCopy_SourceNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	copier := NewKVCopier(server.URL, "test-token", "secret")
	err := copier.Copy("missing/secret", "dst/secret")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestNewKVCopier_DefaultMount(t *testing.T) {
	c := NewKVCopier("http://localhost:8200", "token", "")
	if c.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", c.mount)
	}
}
