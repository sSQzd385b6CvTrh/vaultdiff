package vault

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// KVExporter fetches all key-value pairs under a path and returns them as a flat map.
type KVExporter struct {
	client *api.Client
	mount  string
}

// NewKVExporter creates a new KVExporter.
func NewKVExporter(client *api.Client, mount string) *KVExporter {
	if mount == "" {
		mount = "secret"
	}
	return &KVExporter{client: client, mount: mount}
}

// Export retrieves all secrets under the given path prefix and returns a map of key -> data.
func (e *KVExporter) Export(path string) (map[string]map[string]interface{}, error) {
	listPath := fmt.Sprintf("/v1/%s/metadata/%s", e.mount, path)
	req := e.client.NewRequest(http.MethodGet, listPath)
	req.Params.Set("list", "true")

	resp, err := e.client.RawRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}

	exported := make(map[string]map[string]interface{})
	for _, key := range result.Data.Keys {
		fullKey := fmt.Sprintf("%s/%s", path, key)
		getPath := fmt.Sprintf("/v1/%s/data/%s", e.mount, fullKey)
		greq := e.client.NewRequest(http.MethodGet, getPath)
		gresp, err := e.client.RawRequest(greq)
		if err != nil {
			return nil, fmt.Errorf("get %s failed: %w", fullKey, err)
		}
		defer gresp.Body.Close()
		if gresp.StatusCode != http.StatusOK {
			continue
		}
		var sec struct {
			Data struct {
				Data map[string]interface{} `json:"data"`
			} `json:"data"`
		}
		if err := json.NewDecoder(gresp.Body).Decode(&sec); err == nil {
			exported[fullKey] = sec.Data.Data
		}
	}
	return exported, nil
}
