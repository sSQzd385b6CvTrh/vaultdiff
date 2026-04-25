package vault

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// LatestVersion holds the most recent version number and its data for a KV secret.
type LatestVersion struct {
	Version int
	Data    map[string]string
}

// KVLatestReader fetches the latest version of a KV v2 secret.
type KVLatestReader struct {
	client *api.Client
	mount  string
}

// NewKVLatestReader creates a new KVLatestReader.
func NewKVLatestReader(client *api.Client, mount string) *KVLatestReader {
	if mount == "" {
		mount = "secret"
	}
	return &KVLatestReader{client: client, mount: mount}
}

// GetLatest returns the latest version and data for the given secret path.
func (r *KVLatestReader) GetLatest(ctx context.Context, path string) (*LatestVersion, error) {
	url := fmt.Sprintf("/v1/%s/data/%s", r.mount, path)
	req := r.client.NewRequest(http.MethodGet, url)

	resp, err := r.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for path %s", resp.StatusCode, path)
	}

	var result struct {
		Data struct {
			Data     map[string]interface{} `json:"data"`
			Metadata struct {
				Version int `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	data := make(map[string]string, len(result.Data.Data))
	for k, v := range result.Data.Data {
		data[k] = fmt.Sprintf("%v", v)
	}

	return &LatestVersion{
		Version: result.Data.Metadata.Version,
		Data:    data,
	}, nil
}
