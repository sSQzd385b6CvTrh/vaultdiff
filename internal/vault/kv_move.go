package vault

import (
	"context"
	"fmt"
	"net/http"
)

// KVMover copies a secret from a source path to a destination path,
// then deletes all versions at the source path (destroy).
type KVMover struct {
	copier  *KVCopier
	deleter *KVDeleter
	mount   string
}

// NewKVMover returns a KVMover backed by the given Vault client.
func NewKVMover(client *Client, mount string) *KVMover {
	if mount == "" {
		mount = "secret"
	}
	return &KVMover{
		copier:  NewKVCopier(client, mount),
		deleter: NewKVDeleter(client, mount),
		mount:   mount,
	}
}

// Move copies the secret at srcPath to dstPath and permanently destroys
// the source. Returns an error if either the copy or the destroy fails.
func (m *KVMover) Move(ctx context.Context, srcPath, dstPath string) error {
	if err := m.copier.Copy(ctx, srcPath, dstPath); err != nil {
		return fmt.Errorf("kv move: copy %q -> %q: %w", srcPath, dstPath, err)
	}

	// Retrieve metadata to discover all version numbers to destroy.
	versions, err := m.deleter.allVersions(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("kv move: list versions for %q: %w", srcPath, err)
	}

	if err := m.deleter.Destroy(ctx, srcPath, versions); err != nil {
		return fmt.Errorf("kv move: destroy %q: %w", srcPath, err)
	}
	return nil
}

// allVersions fetches metadata for path and returns a slice of all version numbers.
func (d *KVDeleter) allVersions(ctx context.Context, path string) ([]int, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", d.address, d.mount, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", d.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Versions map[string]interface{} `json:"versions"`
		} `json:"data"`
	}
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	nums := make([]int, 0, len(result.Data.Versions))
	for k := range result.Data.Versions {
		var n int
		if _, err := fmt.Sscanf(k, "%d", &n); err == nil {
			nums = append(nums, n)
		}
	}
	return nums, nil
}
