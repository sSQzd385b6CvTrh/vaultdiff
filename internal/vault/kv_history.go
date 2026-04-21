package vault

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// VersionMeta holds metadata for a single KV secret version.
type VersionMeta struct {
	Version      int
	CreatedTime  time.Time
	DeletionTime time.Time
	Destroyed    bool
}

// KVHistoryReader fetches version history for a KV v2 secret.
type KVHistoryReader struct {
	client *Client
	mount  string
}

// NewKVHistoryReader creates a new KVHistoryReader.
func NewKVHistoryReader(client *Client, mount string) *KVHistoryReader {
	if mount == "" {
		mount = "secret"
	}
	return &KVHistoryReader{client: client, mount: mount}
}

// GetHistory returns all version metadata for the given secret path.
func (r *KVHistoryReader) GetHistory(ctx context.Context, path string) ([]VersionMeta, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", r.client.Address, r.mount, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.client.Token)
	if r.client.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", r.client.Namespace)
	}

	resp, err := r.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Versions map[string]struct {
				CreatedTime  time.Time `json:"created_time"`
				DeletionTime time.Time `json:"deletion_time"`
				Destroyed    bool      `json:"destroyed"`
			} `json:"versions"`
		} `json:"data"`
	}
	if err := decodeJSON(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var history []VersionMeta
	for k, v := range body.Data.Versions {
		var num int
		fmt.Sscanf(k, "%d", &num)
		history = append(history, VersionMeta{
			Version:      num,
			CreatedTime:  v.CreatedTime,
			DeletionTime: v.DeletionTime,
			Destroyed:    v.Destroyed,
		})
	}
	sort.Slice(history, func(i, j int) bool {
		return history[i].Version < history[j].Version
	})
	return history, nil
}
