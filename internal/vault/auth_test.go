package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func authLookupResponse(accessor, displayName string, policies []string, ttl int, renewable bool) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":     accessor,
			"display_name": displayName,
			"policies":     policies,
			"meta":         map[string]string{"env": "prod"},
			"ttl":          ttl,
			"renewable":    renewable,
		},
	}
}

func TestLookupAuth_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(authLookupResponse("abc123", "token-display", []string{"default", "admin"}, 3600, true))
	}))
	defer ts.Close()

	inspector := NewAuthInspector(ts.URL, "test-token")
	info, err := inspector.LookupAuth()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if info.DisplayName != "token-display" {
		t.Errorf("expected display_name token-display, got %s", info.DisplayName)
	}
	if len(info.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(info.Policies))
	}
	if info.TTL != 3600 {
		t.Errorf("expected TTL 3600, got %d", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
}

func TestLookupAuth_Forbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	inspector := NewAuthInspector(ts.URL, "bad-token")
	_, err := inspector.LookupAuth()
	if err == nil {
		t.Fatal("expected error for forbidden response")
	}
}

func TestLookupAuth_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	inspector := NewAuthInspector(ts.URL, "token")
	_, err := inspector.LookupAuth()
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
