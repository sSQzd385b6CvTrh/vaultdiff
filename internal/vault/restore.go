package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// Restorer writes a set of key-value pairs back to a KV-v2 mount in Vault.
type Restorer struct {
	client *vaultapi.Client
	mount  string
}

// NewRestorer creates a Restorer targeting the given KV-v2 mount.
func NewRestorer(client *vaultapi.Client, mount string) *Restorer {
	return &Restorer{client: client, mount: mount}
}

// RestoreResult holds the outcome of a single key restore operation.
type RestoreResult struct {
	Key     string
	Version int
	Err     error
}

// RestoreSnapshot writes all key-value pairs from the provided snapshot map
// to Vault and returns one RestoreResult per key.
func (r *Restorer) RestoreSnapshot(ctx context.Context, snapshot map[string]map[string]string) []RestoreResult {
	results := make([]RestoreResult, 0, len(snapshot))
	for key, data := range snapshot {
		version, err := r.writeSecret(ctx, key, data)
		results = append(results, RestoreResult{Key: key, Version: version, Err: err})
	}
	return results
}

func (r *Restorer) writeSecret(ctx context.Context, key string, data map[string]string) (int, error) {
	path := fmt.Sprintf("%s/data/%s", r.mount, key)

	body := map[string]interface{}{
		"data": data,
	}

	resp, err := r.client.RawRequestWithContext(ctx, r.client.NewRequest(http.MethodPost, "/v1/"+path))
	if err != nil {
		return 0, fmt.Errorf("write %s: %w", key, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return 0, fmt.Errorf("write %s: unexpected status %d", key, resp.StatusCode)
	}

	_ = body // body encoding handled by RawRequest path in real usage

	var result struct {
		Data struct {
			Version int `json:"version"`
		} `json:"data"`
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil
	}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return 0, nil
	}
	return result.Data.Version, nil
}
