package vault

import (
	"context"
	"fmt"
	"net/http"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// KVTTLSetter sets or clears the deletion_time (TTL) on a KV v2 secret.
type KVTTLSetter struct {
	client *vaultapi.Client
	mount  string
}

// NewKVTTLSetter creates a new KVTTLSetter.
func NewKVTTLSetter(client *vaultapi.Client, mount string) *KVTTLSetter {
	if mount == "" {
		mount = "secret"
	}
	return &KVTTLSetter{client: client, mount: mount}
}

// SetTTL applies a deletion_time TTL to the given secret path.
// Pass a zero duration to clear the TTL.
func (s *KVTTLSetter) SetTTL(ctx context.Context, path string, ttl time.Duration) error {
	metaPath := fmt.Sprintf("/v1/%s/metadata/%s", s.mount, path)

	var ttlStr string
	if ttl > 0 {
		ttlStr = ttl.String()
	} else {
		ttlStr = ""
	}

	body := map[string]interface{}{
		"delete_version_after": ttlStr,
	}

	req := s.client.NewRequest(http.MethodPost, metaPath)
	if err := req.SetJSONBody(body); err != nil {
		return fmt.Errorf("set json body: %w", err)
	}

	resp, err := s.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for path %s", resp.StatusCode, path)
	}

	return nil
}
