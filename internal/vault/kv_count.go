package vault

import (
	"context"
	"fmt"
	"strings"
)

// KVCount holds the result of counting keys under a path.
type KVCount struct {
	Path  string
	Count int
	Keys  []string
}

// KVCounter counts the number of keys stored under a KV path.
type KVCounter struct {
	client HTTPClient
	baseURL string
	token   string
	mount   string
}

// NewKVCounter creates a new KVCounter.
func NewKVCounter(client HTTPClient, baseURL, token, mount string) *KVCounter {
	if mount == "" {
		mount = "secret"
	}
	return &KVCounter{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		mount:   mount,
	}
}

// Count returns the number of keys stored under the given path.
func (c *KVCounter) Count(ctx context.Context, path string) (*KVCount, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s?list=true", c.baseURL, c.mount, path)

	resp, err := doRequest(ctx, c.client, "GET", url, c.token, nil)
	if err != nil {
		return nil, fmt.Errorf("kv count request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &KVCount{Path: path, Count: 0, Keys: []string{}}, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %d for path %q", resp.StatusCode, path)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode kv count response: %w", err)
	}

	return &KVCount{
		Path:  path,
		Count: len(result.Data.Keys),
		Keys:  result.Data.Keys,
	}, nil
}
