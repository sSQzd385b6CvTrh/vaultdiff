package vault_test

import (
	"os"
	"testing"

	"github.com/hashicorp/vault/api"

	"github.com/jonathanhope/vaultdiff/internal/vault"
)

func TestKVDiff_Integration(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip("VAULT_ADDR and VAULT_TOKEN must be set for integration tests")
	}

	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.SetToken(token)

	differ := vault.NewKVDiffer(client, "secret")
	result, err := differ.Diff("test/src", "test/dst")
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}

	if result.SourcePath != "test/src" {
		t.Errorf("unexpected source path: %s", result.SourcePath)
	}
	if result.TargetPath != "test/dst" {
		t.Errorf("unexpected target path: %s", result.TargetPath)
	}
}
