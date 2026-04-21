package vault

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
)

// ExpiryInfo holds TTL and expiration metadata for a KV secret.
type ExpiryInfo struct {
	Path        string
	Version     int
	CreatedTime time.Time
	TTL         time.Duration
	ExpiresAt   time.Time
	Expired     bool
}

// KVExpirer checks expiration metadata for KV v2 secrets.
type KVExpirer struct {
	client *api.Client
	mount  string
}

// NewKVExpirer creates a KVExpirer with the given client and mount.
func NewKVExpirer(client *api.Client, mount string) *KVExpirer {
	if mount == "" {
		mount = "secret"
	}
	return &KVExpirer{client: client, mount: mount}
}

// CheckExpiry retrieves metadata for the given path and computes expiry info.
func (e *KVExpirer) CheckExpiry(path string) (*ExpiryInfo, error) {
	url := fmt.Sprintf("/v1/%s/metadata/%s", e.mount, path)
	req := e.client.NewRequest(http.MethodGet, url)

	resp, err := e.client.RawRequest(req)
	if err != nil {
		return nil, fmt.Errorf("expiry check request failed: %w", err)
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
			Versions       map[string]struct {
				CreatedTime  time.Time `json:"created_time"`
				DeletionTime string    `json:"deletion_time"`
				Destroyed    bool      `json:"destroyed"`
			} `json:"versions"`
		} `json:"data"`
	}

	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	ver := result.Data.CurrentVersion
	key := fmt.Sprintf("%d", ver)
	vInfo, ok := result.Data.Versions[key]
	if !ok {
		return nil, fmt.Errorf("version %d not found in metadata", ver)
	}

	info := &ExpiryInfo{
		Path:        path,
		Version:     ver,
		CreatedTime: vInfo.CreatedTime,
	}

	if vInfo.DeletionTime != "" {
		exp, err := time.Parse(time.RFC3339Nano, vInfo.DeletionTime)
		if err == nil {
			info.ExpiresAt = exp
			info.TTL = time.Until(exp)
			info.Expired = time.Now().After(exp)
		}
	}

	return info, nil
}
