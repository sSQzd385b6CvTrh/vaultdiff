package vault

import (
	"context"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// KVPatcher patches (merges) key-value pairs into an existing KVv2 secret.
type KVPatcher struct {
	client *vaultapi.Client
	mount  string
}

// NewKVPatcher creates a new KVPatcher. Defaults mount to "secret".
func NewKVPatcher(client *vaultapi.Client, mount string) *KVPatcher {
	if mount == "" {
		mount = "secret"
	}
	return &KVPatcher{client: client, mount: mount}
}

// Patch merges the provided data into the existing secret at path.
// Uses the KVv2 patch endpoint (HTTP PATCH with merge-patch+json).
func (p *KVPatcher) Patch(ctx context.Context, path string, data map[string]interface{}) error {
	endpoint := fmt.Sprintf("/v1/%s/data/%s", p.mount, path)

	body := map[string]interface{}{
		"data": data,
		"options": map[string]interface{}{},
	}

	req := p.client.NewRequest(http.MethodPatch, endpoint)
	req.Headers.Set("Content-Type", "application/merge-patch+json")
	if err := req.SetJSONBody(body); err != nil {
		return fmt.Errorf("kv patch: encode body: %w", err)
	}

	resp, err := p.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return fmt.Errorf("kv patch: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("kv patch: secret %q not found", path)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("kv patch: unexpected status %d", resp.StatusCode)
	}
	return nil
}
