package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// KVEntry holds the data returned from a KV v2 read.
type KVEntry struct {
	Data    map[string]string
	Version int
}

// KVGetter reads a secret from a KV v2 mount.
type KVGetter struct {
	client *Client
	mount  string
}

// NewKVGetter creates a KVGetter. mount defaults to "secret".
func NewKVGetter(c *Client, mount string) *KVGetter {
	if mount == "" {
		mount = "secret"
	}
	return &KVGetter{client: c, mount: mount}
}

// Get retrieves the latest (or specified) version of a secret.
// Pass version <= 0 to fetch the latest version.
func (g *KVGetter) Get(path string, version int) (*KVEntry, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", g.client.Address, g.mount, path)
	if version > 0 {
		url += fmt.Sprintf("?version=%d", version)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("kv get: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", g.client.Token)
	if g.client.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", g.client.Namespace)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kv get: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("kv get: secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kv get: unexpected status %d", resp.StatusCode)
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
		return nil, fmt.Errorf("kv get: decode: %w", err)
	}

	data := make(map[string]string, len(body.Data.Data))
	for k, v := range body.Data.Data {
		data[k] = fmt.Sprintf("%v", v)
	}
	return &KVEntry{Data: data, Version: body.Data.Metadata.Version}, nil
}
