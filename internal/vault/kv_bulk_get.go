package vault

import (
	"context"
	"fmt"
	"net/http"
)

// BulkGetResult holds the result for a single key in a bulk get operation.
type BulkGetResult struct {
	Key    string
	Data   map[string]interface{}
	Err    error
}

// KVBulkGetter fetches multiple KV secrets in a single operation.
type KVBulkGetter struct {
	client *Client
	mount  string
}

// NewKVBulkGetter creates a new KVBulkGetter.
func NewKVBulkGetter(client *Client, mount string) *KVBulkGetter {
	if mount == "" {
		mount = "secret"
	}
	return &KVBulkGetter{client: client, mount: mount}
}

// Get fetches all provided keys and returns a slice of BulkGetResult.
func (g *KVBulkGetter) Get(ctx context.Context, keys []string) []BulkGetResult {
	results := make([]BulkGetResult, 0, len(keys))
	for _, key := range keys {
		path := fmt.Sprintf("/v1/%s/data/%s", g.mount, key)
		resp, err := g.client.RawGet(ctx, path)
		if err != nil {
			results = append(results, BulkGetResult{Key: key, Err: err})
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			results = append(results, BulkGetResult{Key: key, Err: fmt.Errorf("key not found: %s", key)})
			continue
		}
		if resp.StatusCode != http.StatusOK {
			results = append(results, BulkGetResult{Key: key, Err: fmt.Errorf("unexpected status %d for key %s", resp.StatusCode, key)})
			continue
		}
		var body struct {
			Data struct {
				Data map[string]interface{} `json:"data"`
			} `json:"data"`
		}
		if err := decodeJSON(resp.Body, &body); err != nil {
			results = append(results, BulkGetResult{Key: key, Err: fmt.Errorf("decode error for key %s: %w", key, err)})
			continue
		}
		results = append(results, BulkGetResult{Key: key, Data: body.Data.Data})
	}
	return results
}
