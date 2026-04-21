package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// KVMetadata holds metadata for a KV v2 secret path.
type KVMetadata struct {
	CreatedTime    time.Time         `json:"created_time"`
	UpdatedTime    time.Time         `json:"updated_time"`
	CurrentVersion int               `json:"current_version"`
	OldestVersion  int               `json:"oldest_version"`
	MaxVersions    int               `json:"max_versions"`
	DeleteVersionAfter string        `json:"delete_version_after"`
	CustomMetadata map[string]string `json:"custom_metadata"`
}

// KVMetadataReader fetches metadata for a KV v2 secret.
type KVMetadataReader struct {
	client *http.Client
	address string
	token   string
	mount   string
}

// NewKVMetadataReader creates a new KVMetadataReader.
func NewKVMetadataReader(address, token, mount string) *KVMetadataReader {
	if mount == "" {
		mount = "secret"
	}
	return &KVMetadataReader{
		client:  &http.Client{},
		address: address,
		token:   token,
		mount:   mount,
	}
}

// GetMetadata retrieves the metadata for the given secret path.
func (r *KVMetadataReader) GetMetadata(path string) (*KVMetadata, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", r.address, r.mount, path)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.token)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret metadata not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result struct {
		Data KVMetadata `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result.Data, nil
}
