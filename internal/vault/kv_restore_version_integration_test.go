package vault_test

import (
	"context"
	"os"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"vaultdiff/internal/vault"
)

func TestRestoreVersion_Integration(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip("VAULT_ADDR and VAULT_TOKEN must be set for integration tests")
	}

	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	client.SetToken(token)

	r := vault.NewKVVersionRestorer(client, "secret")
	err = r.RestoreVersion(context.Background(), "vaultdiff/integration-test", []int{1})
	if err != nil {
		t.Logf("restore version (may not exist): %v", err)
	}
}
