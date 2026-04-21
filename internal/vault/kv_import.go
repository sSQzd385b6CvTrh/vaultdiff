package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// KVImportResult holds the result of a single secret import.
type KVImportResult struct {
	Path    string
	Success bool
	Error   error
}

// KVImporter writes multiple secrets from a map into Vault KV v2.
type KVImporter struct {
	client *Client
	mount  string
}

// NewKVImporter creates a new KVImporter. Defaults mount to "secret".
func NewKVImporter(client *Client, mount string) *KVImporter {
	if mount == "" {
		mount = "secret"
	}
	return &KVImporter{client: client, mount: mount}
}

// Import writes each path/data pair from the provided map into Vault.
// It returns a slice of KVImportResult, one per secret path.
func (i *KVImporter) Import(ctx context.Context, secrets map[string]map[string]string) []KVImportResult {
	results := make([]KVImportResult, 0, len(secrets))

	for path, data := range secrets {
		result := KVImportResult{Path: path}

		body := map[string]interface{}{"data": data}
		payload, err := json.Marshal(body)
		if err != nil {
			result.Error = fmt.Errorf("marshal error for %q: %w", path, err)
			results = append(results, result)
			continue
		}

		url := fmt.Sprintf("%s/v1/%s/data/%s", i.client.Address, i.mount, path)
		status, err := i.client.Post(ctx, url, payload)
		if err != nil {
			result.Error = fmt.Errorf("request failed for %q: %w", path, err)
			results = append(results, result)
			continue
		}

		if status != http.StatusOK && status != http.StatusNoContent {
			result.Error = fmt.Errorf("unexpected status %d for %q", status, path)
			results = append(results, result)
			continue
		}

		result.Success = true
		results = append(results, result)
	}

	return results
}
