package vault

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// KVGrepResult holds a matched secret path and the matching key/value pairs.
type KVGrepResult struct {
	Path    string
	Matches map[string]string
}

// KVGrepper searches secret values across multiple paths for a given pattern.
type KVGrepper struct {
	client *Client
	mount  string
}

// NewKVGrepper creates a new KVGrepper.
func NewKVGrepper(client *Client, mount string) *KVGrepper {
	if mount == "" {
		mount = "secret"
	}
	return &KVGrepper{client: client, mount: mount}
}

// Grep searches the secret at path for keys or values matching pattern.
func (g *KVGrepper) Grep(ctx context.Context, path, pattern string, searchKeys bool) (*KVGrepResult, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", g.client.Address, g.mount, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("kv grep: build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", g.client.Token)
	if g.client.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", g.client.Namespace)
	}

	resp, err := g.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kv grep: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("kv grep: path %q not found", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kv grep: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := decodeJSON(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("kv grep: decode: %w", err)
	}

	matches := make(map[string]string)
	lower := strings.ToLower(pattern)
	for k, v := range body.Data.Data {
		val := fmt.Sprintf("%v", v)
		if searchKeys && strings.Contains(strings.ToLower(k), lower) {
			matches[k] = val
		} else if !searchKeys && strings.Contains(strings.ToLower(val), lower) {
			matches[k] = val
		}
	}

	if len(matches) == 0 {
		return nil, nil
	}
	return &KVGrepResult{Path: path, Matches: matches}, nil
}
