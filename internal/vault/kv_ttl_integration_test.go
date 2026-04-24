package vault

import (
	"os"
	"testing"
)

// TestGetTTL_Integration runs against a live Vault instance.
// Set VAULT_ADDR, VAULT_TOKEN, and VAULT_INTEGRATION_PATH to enable.
func TestGetTTL_Integration(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	path := os.Getenv("VAULT_INTEGRATION_PATH")

	if addr == "" || token == "" || path == "" {
		t.Skip("skipping integration test: VAULT_ADDR, VAULT_TOKEN, VAULT_INTEGRATION_PATH not set")
	}

	client, err := NewClient(addr, token, "")
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	reader := NewKVTTLReader(client.API(), "secret")
	info, err := reader.GetTTL(path)
	if err != nil {
		t.Fatalf("GetTTL error: %v", err)
	}

	t.Logf("Path: %s, HasTTL: %v, Remaining: %v", info.Path, info.HasTTL, info.Remaining)
}
