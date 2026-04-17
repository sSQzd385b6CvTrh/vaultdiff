package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// KVLister lists keys under a KV v2 path.
type KVLister struct {
	address string
	token   string
	mount   string
	client  *http.Client
}

// NewKVLister creates a new KVLister.
func NewKVLister(address, token, mount string) *KVLister {
	if mount == "" {
		mount = "secret"
	}
	return &KVLister{
		address: address,
		token:   token,
		mount:   mount,
		client:  &http.Client{},
	}
}

// ListKeys returns all keys under the given path prefix.
func (l *KVLister) ListKeys(path string) ([]string, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s?list=true", l.address, l.mount, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", l.token)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if result.Data.Keys == nil {
		return nil, fmt.Errorf("missing keys in response")
	}
	return result.Data.Keys, nil
}
