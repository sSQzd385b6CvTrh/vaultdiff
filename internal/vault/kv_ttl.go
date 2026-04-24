package vault

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

// TTLInfo holds the TTL metadata for a KV secret.
type TTLInfo struct {
	Path        string
	DeletionTime time.Time
	HasTTL      bool
	Remaining   time.Duration
}

// KVTTLReader reads TTL information from KV v2 secret metadata.
type KVTTLReader struct {
	client *api.Client
	mount  string
}

// NewKVTTLReader creates a new KVTTLReader.
func NewKVTTLReader(client *api.Client, mount string) *KVTTLReader {
	if mount == "" {
		mount = "secret"
	}
	return &KVTTLReader{client: client, mount: mount}
}

// GetTTL returns TTL information for the given secret path.
func (r *KVTTLReader) GetTTL(path string) (*TTLInfo, error) {
	metaPath := fmt.Sprintf("%s/metadata/%s", r.mount, path)
	secret, err := r.client.Logical().Read(metaPath)
	if err != nil {
		return nil, fmt.Errorf("reading metadata for %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found: %s", path)
	}

	info := &TTLInfo{Path: path}

	raw, ok := secret.Data["delete_version_after"]
	if !ok || raw == nil || raw == "0s" {
		return info, nil
	}

	delStr, _ := raw.(string)
	if delStr == "" || delStr == "0s" {
		return info, nil
	}

	d, err := time.ParseDuration(delStr)
	if err != nil {
		return nil, fmt.Errorf("parsing delete_version_after %q: %w", delStr, err)
	}

	info.HasTTL = true
	info.DeletionTime = time.Now().Add(d)
	info.Remaining = d
	return info, nil
}
