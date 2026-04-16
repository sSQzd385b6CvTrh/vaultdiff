package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func leaseResponse(id string, renewable bool, duration int) map[string]interface{} {
	return map[string]interface{}{
		"id":             id,
		"renewable":      renewable,
		"lease_duration": duration,
	}
}

func newLeaseServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestLookupLease_Success(t *testing.T) {
	payload := leaseResponse("auth/token/create/abc123", true, 3600)
	srv := newLeaseServer(t, http.StatusOK, payload)
	defer srv.Close()

	inspector := NewLeaseInspector(srv.URL, "test-token")
	info, err := inspector.Lookup("auth/token/create/abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.LeaseID != "auth/token/create/abc123" {
		t.Errorf("expected lease ID, got %s", info.LeaseID)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
	if info.LeaseDuration.Seconds() != 3600 {
		t.Errorf("expected 3600s, got %v", info.LeaseDuration)
	}
}

func TestLookupLease_NotFound(t *testing.T) {
	srv := newLeaseServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	inspector := NewLeaseInspector(srv.URL, "test-token")
	_, err := inspector.Lookup("missing/lease")
	if err == nil {
		t.Fatal("expected error for missing lease")
	}
}

func TestLookupLease_ServerError(t *testing.T) {
	srv := newLeaseServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	inspector := NewLeaseInspector(srv.URL, "test-token")
	_, err := inspector.Lookup("some/lease")
	if err == nil {
		t.Fatal("expected error on server failure")
	}
}
