package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func sealStatusResponse(sealed bool) map[string]interface{} {
	return map[string]interface{}{
		"sealed":      sealed,
		"initialized": true,
		"progress":    0,
		"t":           3,
		"n":           5,
		"version":     "1.15.0",
	}
}

func newSealServer(t *testing.T, payload map[string]interface{}, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func TestSealStatus_Unsealed(t *testing.T) {
	server := newSealServer(t, sealStatusResponse(false), http.StatusOK)
	defer server.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, _ := vaultapi.NewClient(cfg)

	checker := NewSealChecker(client)
	status, err := checker.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Sealed {
		t.Error("expected vault to be unsealed")
	}
	if !status.Initialized {
		t.Error("expected vault to be initialized")
	}
	if status.Threshold != 3 {
		t.Errorf("expected threshold 3, got %d", status.Threshold)
	}
	if status.Total != 5 {
		t.Errorf("expected total 5, got %d", status.Total)
	}
	if status.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", status.Version)
	}
}

func TestSealStatus_Sealed(t *testing.T) {
	server := newSealServer(t, sealStatusResponse(true), http.StatusOK)
	defer server.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, _ := vaultapi.NewClient(cfg)

	checker := NewSealChecker(client)
	status, err := checker.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Sealed {
		t.Error("expected vault to be sealed")
	}
}
