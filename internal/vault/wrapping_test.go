package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func wrappingLookupResponse(info WrappingInfo) map[string]interface{} {
	return map[string]interface{}{"data": info}
}

func TestLookupWrapping_Success(t *testing.T) {
	expected := WrappingInfo{
		Token:        "s.abc123",
		Accessor:     "acc-xyz",
		TTL:          300,
		CreationTime: "2024-01-01T00:00:00Z",
		CreationPath: "secret/data/foo",
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/sys/wrapping/lookup" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(wrappingLookupResponse(expected))
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "root"}
	inspector := NewWrappingInspector(c)
	info, err := inspector.Lookup("s.abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Token != expected.Token {
		t.Errorf("expected token %q, got %q", expected.Token, info.Token)
	}
	if info.TTL != expected.TTL {
		t.Errorf("expected TTL %d, got %d", expected.TTL, info.TTL)
	}
	if info.CreationPath != expected.CreationPath {
		t.Errorf("expected creation path %q, got %q", expected.CreationPath, info.CreationPath)
	}
}

func TestLookupWrapping_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "root"}
	inspector := NewWrappingInspector(c)
	_, err := inspector.Lookup("s.expired")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestLookupWrapping_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "root"}
	inspector := NewWrappingInspector(c)
	_, err := inspector.Lookup("s.bad")
	if err == nil {
		t.Fatal("expected error for 500")
	}
}
