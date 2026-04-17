package vault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// KVDeleter deletes a secret version from a KV v2 mount.
type KVDeleter struct {
	client *api.Client
	mount  string
}

// NewKVDeleter creates a new KVDeleter.
func NewKVDeleter(client *api.Client, mount string) *KVDeleter {
	if mount == "" {
		mount = "secret"
	}
	return &KVDeleter{client: client, mount: mount}
}

// Delete soft-deletes the specified versions of a secret.
func (d *KVDeleter) Delete(path string, versions []int) error {
	body := map[string]interface{}{"versions": versions}
	resp, err := d.client.Logical().WriteWithContext(
		d.client.CloneConfig().Context,
		fmt.Sprintf("%s/delete/%s", d.mount, path),
		body,
	)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	if resp != nil && resp.Data != nil {
		if warn, ok := resp.Data["warnings"]; ok {
			return fmt.Errorf("vault warning: %v", warn)
		}
	}
	return nil
}

// Destroy permanently destroys the specified versions of a secret.
func (d *KVDeleter) Destroy(path string, versions []int) error {
	body := map[string]interface{}{"versions": versions}
	req := d.client.NewRequest(http.MethodPut,mt.Sprintf("/v1/%s/destroy/%s", d.mount, path))
	if err := req.SetJSONBody(body); err != nil {
		return fmt.Errorf("failed to encode body: %w", err)
	}
	rawResp, err := d.client.RawRequest(req)
	if err != nil {
		return fmt.Errorf("destroy request failed: %w", err)
	}
	defer rawResp.Body.Close()
	if rawResp.StatusCode != http.StatusNoContent && rawResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", rawResp.StatusCode)
	}
	return nil
}
