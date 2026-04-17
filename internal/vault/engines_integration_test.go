//go:build integration
// +build integration

package vault

import (
	"os"
	"testing"
)

func TestListEngines_Integration(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		t.Skip("VAULT_ADDR not set, skipping integration test")
	}
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		t.Skip("VAULT_TOKEN not set, skipping integration test")
	}

	c, err := NewClient(addr, token, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	lister := NewSecretEngineLister(c)
	result, err := lister.ListEngines()
	if err != nil {
		t.Fatalf("ListEngines failed: %v", err)
	}
	if len(result.Engines) == 0 {
		t.Error("expected at least one mounted engine")
	}
	for _, e := range result.Engines {
		t.Logf("engine: path=%s type=%s", e.Path, e.Type)
	}
}
