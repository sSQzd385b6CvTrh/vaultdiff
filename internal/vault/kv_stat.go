package vault

import (
	"context"
	"fmt"
	"net/http"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretStat holds statistical information about a KV secret.
type SecretStat struct {
	Path         string
	CurrentVersion int
	TotalVersions  int
	CreatedTime    time.Time
	UpdatedTime    time.Time
	Destroyed      bool
	Deleted        bool
}

// KVStatter retrieves stat information for a KV v2 secret path.
type KVStatter struct {
	client *vaultapi.Client
	mount  string
}

// NewKVStatter creates a new KVStatter with the given client and mount.
func NewKVStatter(client *vaultapi.Client, mount string) *KVStatter {
	if mount == "" {
		mount = "secret"
	}
	return &KVStatter{client: client, mount: mount}
}

// Stat returns statistical metadata for the secret at path.
func (s *KVStatter) Stat(ctx context.Context, path string) (*SecretStat, error) {
	url := fmt.Sprintf("/v1/%s/metadata/%s", s.mount, path)
	req := s.client.NewRequest(http.MethodGet, url)

	resp, err := s.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("stat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for path %s", resp.StatusCode, path)
	}

	var result struct {
		Data struct {
			CurrentVersion int                       `json:"current_version"`
			OldestVersion  int                       `json:"oldest_version"`
			CreatedTime    time.Time                 `json:"created_time"`
			UpdatedTime    time.Time                 `json:"updated_time"`
			Versions       map[string]struct {
				Destroyed   bool       `json:"destroyed"`
				DeletionTime string    `json:"deletion_time"`
			} `json:"versions"`
		} `json:"data"`
	}

	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to decode stat response: %w", err)
	}

	d := result.Data
	current := d.CurrentVersion
	v, hasVersion := d.Versions[fmt.Sprintf("%d", current)]

	stat := &SecretStat{
		Path:           path,
		CurrentVersion: current,
		TotalVersions:  len(d.Versions),
		CreatedTime:    d.CreatedTime,
		UpdatedTime:    d.UpdatedTime,
	}
	if hasVersion {
		stat.Destroyed = v.Destroyed
		stat.Deleted = v.DeletionTime != ""
	}
	return stat, nil
}
