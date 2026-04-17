package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func engineListResponse() map[string]interface{} {
	return map[string]interface{}{
		"secret/": map[string]interface{}{
			"type":        "kv",
			"description": "key/value store",
			"accessor":    "kv_abc123",
		},
		"pki/": map[string]interface{}{
			"type":        "pki",
			"description": "PKI engine",
			"accessor":    "pki_def456",
		},
		"request_id": "ignored-string",
	}
}

func TestListEngines_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(engineListResponse())
	}))
	defer ts.Close()

	c, _ := NewClient(ts.URL, "token", "")
	lister := NewSecretEngineLister(c)
	result, err := lister.ListEngines()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Engines) != 2 {
		t.Fatalf("expected 2 engines, got %d", len(result.Engines))
	}
	if result.Engines[0].Path != "pki/" {
		t.Errorf("expected sorted first path pki/, got %s", result.Engines[0].Path)
	}
}

func TestListEngines_Forbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	c, _ := NewClient(ts.URL, "token", "")
	lister := NewSecretEngineLister(c)
	_, err := lister.ListEngines()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListEngines_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c, _ := NewClient(ts.URL, "token", "")
	lister := NewSecretEngineLister(c)
	_, err := lister.ListEngines()
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}
