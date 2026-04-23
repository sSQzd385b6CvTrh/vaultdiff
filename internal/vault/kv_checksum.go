package vault

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/snyk/vaultdiff/internal/vault/api"
)

// KVChecksum holds the SHA-256 checksum of a secret's data at a given version.
type KVChecksum struct {
	Path    string
	Version int
	Sum     string
}

// KVChecksummer computes deterministic checksums for KV secret versions.
type KVChecksummer struct {
	client *api.Client
	mount  string
}

// NewKVChecksummer returns a KVChecksummer using the provided client.
// mount defaults to "secret" if empty.
func NewKVChecksummer(client *api.Client, mount string) *KVChecksummer {
	if mount == "" {
		mount = "secret"
	}
	return &KVChecksummer{client: client, mount: mount}
}

// Checksum fetches the specified version of a KV secret and returns its SHA-256
// checksum computed over the canonically sorted JSON representation of the data.
func (c *KVChecksummer) Checksum(path string, version int) (*KVChecksum, error) {
	url := fmt.Sprintf("/v1/%s/data/%s?version=%d", c.mount, path, version)
	resp, err := c.client.RawRequestWithContext(nil, c.client.NewRequest(http.MethodGet, url))
	if err != nil {
		return nil, fmt.Errorf("checksum request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s version %d", path, version)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, path)
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	sum, err := checksumData(result.Data.Data)
	if err != nil {
		return nil, fmt.Errorf("checksum compute failed: %w", err)
	}

	return &KVChecksum{Path: path, Version: version, Sum: sum}, nil
}

// checksumData produces a deterministic SHA-256 hex digest over sorted JSON.
func checksumData(data map[string]interface{}) (string, error) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ordered := make([]interface{}, 0, len(keys))
	for _, k := range keys {
		ordered = append(ordered, []interface{}{k, data[k]})
	}

	b, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h), nil
}
