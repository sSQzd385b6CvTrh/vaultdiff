package vault

import (
	"context"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// KVExistsResult holds the result of a KV existence check.
type KVExistsResult struct {
	Path    string
	Exists  bool
	Version int
	Deleted bool
	Destroyed bool
}

// KVExistenceChecker checks whether a KV secret path exists.
type KVExistenceChecker struct {
	client *vaultapi.Client
	mount  string
}

// NewKVExistenceChecker creates a new KVExistenceChecker.
func NewKVExistenceChecker(client *vaultapi.Client, mount string) *KVExistenceChecker {
	if mount == "" {
		mount = "secret"
	}
	return &KVExistenceChecker{client: client, mount: mount}
}

// Check returns a KVExistsResult for the given secret path.
func (e *KVExistenceChecker) Check(ctx context.Context, path string) (*KVExistsResult, error) {
	apiPath := fmt.Sprintf("/v1/%s/data/%s", e.mount, path)
	req := e.client.NewRequest(http.MethodGet, apiPath)

	resp, err := e.client.RawRequestWithContext(ctx, req)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return &KVExistsResult{Path: path, Exists: false}, nil
		}
		return nil, fmt.Errorf("existence check failed for %q: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &KVExistsResult{Path: path, Exists: false}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for path %q", resp.StatusCode, path)
	}

	var result struct {
		Data struct {
			Metadata struct {
				Version   int  `json:"version"`
				Deleted   bool `json:"deletion_time"`
				Destroyed bool `json:"destroyed"`
			} `json:"metadata"`
		} `json:"data"`
	}

	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response for %q: %w", path, err)
	}

	return &KVExistsResult{
		Path:      path,
		Exists:    true,
		Version:   result.Data.Metadata.Version,
		Deleted:   result.Data.Metadata.Deleted,
		Destroyed: result.Data.Metadata.Destroyed,
	}, nil
}
