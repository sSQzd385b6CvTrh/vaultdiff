package vault

import (
	"context"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// KVUndeleter restores previously soft-deleted versions of a KV v2 secret.
type KVUndeleter struct {
	client *vaultapi.Client
	mount  string
}

// NewKVUndeleter creates a new KVUndeleter. If mount is empty, "secret" is used.
func NewKVUndeleter(client *vaultapi.Client, mount string) *KVUndeleter {
	if mount == "" {
		mount = "secret"
	}
	return &KVUndeleter{client: client, mount: mount}
}

// Undelete restores the specified versions of the secret at path.
func (u *KVUndeleter) Undelete(ctx context.Context, path string, versions []int) error {
	if len(versions) == 0 {
		return fmt.Errorf("at least one version must be specified")
	}

	body := map[string]interface{}{
		"versions": versions,
	}

	endpoint := fmt.Sprintf("/v1/%s/undelete/%s", u.mount, path)
	req := u.client.NewRequest(http.MethodPost, endpoint)
	if err := req.SetJSONBody(body); err != nil {
		return fmt.Errorf("setting request body: %w", err)
	}

	resp, err := u.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return fmt.Errorf("undelete request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("secret path %q not found", path)
	case http.StatusForbidden:
		return fmt.Errorf("permission denied for path %q", path)
	default:
		return fmt.Errorf("unexpected status %d for path %q", resp.StatusCode, path)
	}
}
