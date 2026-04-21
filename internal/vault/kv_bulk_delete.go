package vault

import (
	"context"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// BulkDeleteResult holds the outcome for a single path in a bulk delete operation.
type BulkDeleteResult struct {
	Path    string
	Deleted bool
	Err     error
}

// KVBulkDeleter deletes multiple KV secrets in a single operation.
type KVBulkDeleter struct {
	client *vaultapi.Client
	mount  string
}

// NewKVBulkDeleter creates a new KVBulkDeleter.
func NewKVBulkDeleter(client *vaultapi.Client, mount string) *KVBulkDeleter {
	if mount == "" {
		mount = "secret"
	}
	return &KVBulkDeleter{client: client, mount: mount}
}

// DeleteAll deletes all provided secret paths and returns per-path results.
func (b *KVBulkDeleter) DeleteAll(ctx context.Context, paths []string) []BulkDeleteResult {
	results := make([]BulkDeleteResult, 0, len(paths))
	for _, p := range paths {
		result := BulkDeleteResult{Path: p}
		url := fmt.Sprintf("/v1/%s/data/%s", b.mount, p)
		req := b.client.NewRequest(http.MethodDelete, url)
		resp, err := b.client.RawRequestWithContext(ctx, req)
		if err != nil {
			result.Err = fmt.Errorf("delete %q: %w", p, err)
			results = append(results, result)
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			result.Deleted = true
		} else if resp.StatusCode == http.StatusNotFound {
			result.Err = fmt.Errorf("delete %q: secret not found", p)
		} else {
			result.Err = fmt.Errorf("delete %q: unexpected status %d", p, resp.StatusCode)
		}
		results = append(results, result)
	}
	return results
}
