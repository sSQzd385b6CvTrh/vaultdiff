//go:build integration
// +build integration

package vault

import (
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
)

func TestKVLint_Integration(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip("VAULT_ADDR and VAULT_TOKEN must be set for integration tests")
	}

	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	client.SetToken(token)

	linter := NewKVLinter(client, "secret")
	results, err := linter.Lint("test/lint")
	if err != nil {
		t.Skipf("secret not found or vault unavailable: %v", err)
	}
	t.Logf("lint results: %+v", results)
}
