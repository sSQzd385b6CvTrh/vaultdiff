package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func tokenLookupResponse(accessor, displayName string, policies []string, ttl int, renewable bool, expireTime string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":     accessor,
			"display_name": displayName,
			"policies":     policies,
			"ttl":          ttl,
			"renewable":    renewable,
			"expire_time":  expireTime,
		},
	}
}

func TestLookupSelf_Success(t *testing.T) {
	expire := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	payload := tokenLookupResponse("abc123", "token-test", []string{"default", "dev"}, 3600, true, expire)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token/lookup-self" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	inspector := NewTokenInspector(ts.URL, "test-token")
	info, err := inspector.LookupSelf()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if info.DisplayName != "token-test" {
		t.Errorf("expected display_name token-test, got %s", info.DisplayName)
	}
	if len(info.Policies) != 2 || info.Policies[1] != "dev" {
		t.Errorf("unexpected policies: %v", info.Policies)
	}
	if info.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected token to be renewable")
	}
	if info.ExpireTime == nil {
		t.Error("expected expire time to be set")
	}
}

func TestLookupSelf_Forbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	inspector := NewTokenInspector(ts.URL, "bad-token")
	_, err := inspector.LookupSelf()
	if err == nil {
		t.Fatal("expected error for forbidden response")
	}
}

func TestLookupSelf_NoExpireTime(t *testing.T) {
	payload := tokenLookupResponse("xyz", "root", []string{"root"}, 0, false, "")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer ts.Close()

	inspector := NewTokenInspector(ts.URL, "root-token")
	info, err := inspector.LookupSelf()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ExpireTime != nil {
		t.Errorf("expected nil expire time, got %v", info.ExpireTime)
	}
}
