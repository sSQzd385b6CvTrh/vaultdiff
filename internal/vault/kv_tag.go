package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/drew/vaultdiff/internal/vault/api"
)

// KVTagger manages custom metadata tags on KV v2 secrets.
type KVTagger struct {
	client *api.Client
	mount  string
}

// NewKVTagger returns a KVTagger for the given mount (default "secret").
func NewKVTagger(client *api.Client, mount string) *KVTagger {
	if mount == "" {
		mount = "secret"
	}
	return &KVTagger{client: client, mount: mount}
}

// SetTags writes custom_metadata tags for the given secret path.
func (t *KVTagger) SetTags(path string, tags map[string]string) error {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", t.client.Address(), t.mount, path)

	body := map[string]interface{}{
		"custom_metadata": tags,
	}
	resp, err := t.client.RawPost(url, body)
	if err != nil {
		return fmt.Errorf("set tags request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status setting tags: %d", resp.StatusCode)
	}
	return nil
}

// GetTags retrieves custom_metadata tags for the given secret path.
func (t *KVTagger) GetTags(path string) (map[string]string, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", t.client.Address(), t.mount, path)

	resp, err := t.client.RawGet(url)
	if err != nil {
		return nil, fmt.Errorf("get tags request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret path not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status getting tags: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var result struct {
		Data struct {
			CustomMetadata map[string]string `json:"custom_metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if result.Data.CustomMetadata == nil {
		return map[string]string{}, nil
	}
	return result.Data.CustomMetadata, nil
}
