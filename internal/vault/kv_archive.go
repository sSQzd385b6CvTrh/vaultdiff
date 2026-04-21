package vault

import (
	"context"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// ArchiveEntry holds a key and all its versioned data.
type ArchiveEntry struct {
	Key      string
	Versions map[string]map[string]interface{}
}

// KVArchiver fetches all versions of a secret and returns an archive.
type KVArchiver struct {
	client *vaultapi.Client
	mount  string
}

// NewKVArchiver creates a new KVArchiver.
func NewKVArchiver(client *vaultapi.Client, mount string) *KVArchiver {
	if mount == "" {
		mount = "secret"
	}
	return &KVArchiver{client: client, mount: mount}
}

// Archive retrieves all available versions of a KV v2 secret.
func (a *KVArchiver) Archive(ctx context.Context, path string) (*ArchiveEntry, error) {
	metaPath := fmt.Sprintf("/v1/%s/metadata/%s", a.mount, path)
	req := a.client.NewRequest(http.MethodGet, metaPath)
	resp, err := a.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("archive metadata request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching metadata", resp.StatusCode)
	}
	var meta struct {
		Data struct {
			Versions map[string]struct {
				DeletionTime  string `json:"deletion_time"`
				Destroyed     bool   `json:"destroyed"`
			} `json:"versions"`
		} `json:"data"`
	}
	if err := resp.DecodeJSON(&meta); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}
	entry := &ArchiveEntry{
		Key:      path,
		Versions: make(map[string]map[string]interface{}),
	}
	for ver, info := range meta.Data.Versions {
		if info.Destroyed {
			continue
		}
		dataPath := fmt.Sprintf("/v1/%s/data/%s?version=%s", a.mount, path, ver)
		dreq := a.client.NewRequest(http.MethodGet, dataPath)
		dresp, err := a.client.RawRequestWithContext(ctx, dreq)
		if err != nil || dresp.StatusCode != http.StatusOK {
			continue
		}
		var body struct {
			Data struct {
				Data map[string]interface{} `json:"data"`
			} `json:"data"`
		}
		_ = dresp.DecodeJSON(&body)
		dresp.Body.Close()
		entry.Versions[ver] = body.Data.Data
	}
	return entry, nil
}
