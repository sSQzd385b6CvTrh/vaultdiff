package vault

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretSnapshot represents a point-in-time capture of a secret's data.
type SecretSnapshot struct {
	Path      string
	Version   int
	Data      map[string]string
	CreatedAt time.Time
	Deleted   bool
}

// Snapshotter captures secret snapshots from Vault KV v2.
type Snapshotter struct {
	client *vaultapi.Client
	mount  string
}

// NewSnapshotter creates a Snapshotter for the given KV mount.
func NewSnapshotter(client *vaultapi.Client, mount string) *Snapshotter {
	if mount == "" {
		mount = "secret"
	}
	return &Snapshotter{client: client, mount: mount}
}

// Capture fetches the specified version of a secret and returns a snapshot.
// If version is 0, the latest version is fetched.
func (s *Snapshotter) Capture(ctx context.Context, path string, version int) (*SecretSnapshot, error) {
	var endpoint string
	if version > 0 {
		endpoint = fmt.Sprintf("/v1/%s/data/%s?version=%d", s.mount, path, version)
	} else {
		endpoint = fmt.Sprintf("/v1/%s/data/%s", s.mount, path)
	}

	req := s.client.NewRequest("GET", endpoint)
	resp, err := s.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("snapshot request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("secret not found: %s", path)
	}

	var result map[string]interface{}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("decoding snapshot response: %w", err)
	}

	dataBlock, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing data block in response")
	}

	meta, _ := dataBlock["metadata"].(map[string]interface{})
	kvData, _ := dataBlock["data"].(map[string]interface{})

	snap := &SecretSnapshot{
		Path:    path,
		Version: version,
		Data:    make(map[string]string),
	}

	if meta != nil {
		if v, ok := meta["version"].(float64); ok {
			snap.Version = int(v)
		}
		if d, ok := meta["destroyed"].(bool); ok {
			snap.Deleted = d
		}
		if ts, ok := meta["created_time"].(string); ok {
			if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
				snap.CreatedAt = t
			}
		}
	}

	for k, v := range kvData {
		if sv, ok := v.(string); ok {
			snap.Data[k] = sv
		}
	}

	return snap, nil
}
