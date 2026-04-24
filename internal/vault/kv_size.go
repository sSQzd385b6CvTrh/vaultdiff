package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// KVSizeResult holds the computed size information for a KV secret.
type KVSizeResult struct {
	Path       string
	Version    int
	KeyCount   int
	TotalBytes int
}

// KVSizer fetches a KV secret and computes its size metadata.
type KVSizer struct {
	address string
	token   string
	mount   string
}

// NewKVSizer returns a new KVSizer. If mount is empty, "secret" is used.
func NewKVSizer(address, token, mount string) *KVSizer {
	if mount == "" {
		mount = "secret"
	}
	return &KVSizer{address: address, token: token, mount: mount}
}

// Measure fetches the given path at the specified version and returns size info.
// Pass version <= 0 to fetch the latest version.
func (s *KVSizer) Measure(path string, version int) (*KVSizeResult, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", s.address, s.mount, path)
	if version > 0 {
		url = fmt.Sprintf("%s?version=%d", url, version)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("kv size: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", s.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kv size: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("kv size: secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kv size: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Data     map[string]interface{} `json:"data"`
			Metadata struct {
				Version int `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("kv size: decode response: %w", err)
	}

	totalBytes := 0
	for k, v := range body.Data.Data {
		totalBytes += len(k)
		if s, ok := v.(string); ok {
			totalBytes += len(s)
		}
	}

	return &KVSizeResult{
		Path:       path,
		Version:    body.Data.Metadata.Version,
		KeyCount:   len(body.Data.Data),
		TotalBytes: totalBytes,
	}, nil
}
