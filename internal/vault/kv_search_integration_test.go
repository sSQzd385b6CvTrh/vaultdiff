package vault_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"

	"github.com/jonnylangefeld/vaultdiff/internal/vault"
)

func TestKVSearch_Integration(t *testing.T) {
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

	searcher := vault.NewKVSearcher(client, "secret")
	results, err := searcher.Search(context.Background(), "", "password")
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	t.Logf("found %d secrets with 'password' key", len(results))
}
