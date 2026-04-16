//go:build integration
// +build integration

package vault

import (
	"os"
	"testing"
)

// TestRenew_Integration tests token renewal against a real Vault instance.
// Run with: go test -tags integration ./internal/vault/...
func TestRenew_Integration(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		t.Skip("VAULT_ADDR not set, skipping integration test")
	}
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		t.Skip("VAULT_TOKEN not set, skipping integration test")
	}

	renewer := NewTokenRenewer(addr, token)
	result, err := renewer.Renew()
	if err != nil {
		t.Fatalf("Renew failed: %v", err)
	}
	if result.ClientToken == "" {
		t.Error("expected non-empty client token")
	}
	t.Logf("Renewed token, lease duration: %s, renewable: %v", result.LeaseDuration, result.Renewable)
}
