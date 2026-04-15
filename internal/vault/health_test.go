package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func healthResponse(initialized, sealed, standby bool, version string) map[string]interface{} {
	return map[string]interface{}{
		"initialized":  initialized,
		"sealed":       sealed,
		"standby":      standby,
		"version":      version,
		"cluster_name": "vault-cluster",
		"cluster_id":   "abc-123",
	}
}

func newHealthServer(t *testing.T, status int, body map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
}

func TestHealthCheck_Healthy(t *testing.T) {
	body := healthResponse(true, false, false, "1.15.0")
	srv := newHealthServer(t, http.StatusOK, body)
	defer srv.Close()

	checker := NewHealthChecker(srv.URL)
	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Initialized {
		t.Error("expected initialized to be true")
	}
	if status.Sealed {
		t.Error("expected sealed to be false")
	}
	if status.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", status.Version)
	}
}

func TestHealthCheck_Sealed(t *testing.T) {
	body := healthResponse(true, true, false, "1.15.0")
	srv := newHealthServer(t, http.StatusServiceUnavailable, body)
	defer srv.Close()

	checker := NewHealthChecker(srv.URL)
	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Sealed {
		t.Error("expected sealed to be true")
	}
}

func TestHealthCheck_Standby(t *testing.T) {
	body := healthResponse(true, false, true, "1.15.0")
	srv := newHealthServer(t, http.StatusOK, body)
	defer srv.Close()

	checker := NewHealthChecker(srv.URL)
	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Standby {
		t.Error("expected standby to be true")
	}
}

func TestNewHealthChecker_Address(t *testing.T) {
	checker := NewHealthChecker("http://vault.example.com:8200")
	if checker.address != "http://vault.example.com:8200" {
		t.Errorf("unexpected address: %s", checker.address)
	}
}
