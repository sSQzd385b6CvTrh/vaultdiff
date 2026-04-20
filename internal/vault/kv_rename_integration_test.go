package vault

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
)

func TestKVRename_Integration(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip("skipping integration test: VAULT_ADDR or VAULT_TOKEN not set")
	}

	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.SetToken(token)

	ctx := context.Background()
	mount := "secret"
	src := "vaultdiff/rename-src"
	dst := "vaultdiff/rename-dst"

	// Seed source secret
	_, err = client.Logical().WriteWithContext(ctx,
		mount+"/data/"+src,
		map[string]interface{}{"data": map[string]interface{}{"hello": "world"}},
	)
	if err != nil {
		t.Fatalf("setup: write source: %v", err)
	}

	renamer := NewKVRenamer(client, mount)
	if err := renamer.Rename(ctx, src, dst); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Verify destination exists
	secret, err := client.Logical().ReadWithContext(ctx, mount+"/data/"+dst)
	if err != nil || secret == nil {
		t.Fatalf("expected destination secret to exist after rename")
	}

	// Cleanup
	client.Logical().DeleteWithContext(ctx, mount+"/metadata/"+dst) //nolint
}
