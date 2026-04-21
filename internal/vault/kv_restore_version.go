package vault

import (
	"context"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// KVVersionRestorer restores a specific soft-deleted version of a KV v2 secret.
type KVVersionRestorer struct {
	client *vaultapi.Client
	mount  string
}

// NewKVVersionRestorer creates a new KVVersionRestorer.
func NewKVVersionRestorer(client *vaultapi.Client, mount string) *KVVersionRestorer {
	if mount == "" {
		mount = "secret"
	}
	return &KVVersionRestorer{client: client, mount: mount}
}

// RestoreVersion undeletes the given versions of a KV v2 secret at path.
func (r *KVVersionRestorer) RestoreVersion(ctx context.Context, path string, versions []int) error {
	if len(versions) == 0 {
		return fmt.Errorf("no versions specified")
	}

	body := map[string]interface{}{
		"versions": versions,
	}

	endpoint := fmt.Sprintf("/v1/%s/undelete/%s", r.mount, path)
	req := r.client.NewRequest(http.MethodPost, endpoint)
	if err := req.SetJSONBody(body); err != nil {
		return fmt.Errorf("set body: %w", err)
	}

	resp, err := r.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return fmt.Errorf("undelete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("secret path not found: %s", path)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
