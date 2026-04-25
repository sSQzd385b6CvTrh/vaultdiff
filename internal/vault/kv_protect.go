package vault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// KVProtector manages write-protection flags on KV secrets via custom metadata.
type KVProtector struct {
	client *api.Client
	mount  string
}

// NewKVProtector returns a KVProtector targeting the given mount.
func NewKVProtector(client *api.Client, mount string) *KVProtector {
	if mount == "" {
		mount = "secret"
	}
	return &KVProtector{client: client, mount: mount}
}

// Protect sets the custom metadata key "protected" to "true" on the given path.
func (p *KVProtector) Protect(path string) error {
	return p.setProtected(path, "true")
}

// Unprotect removes the protection flag by setting "protected" to "false".
func (p *KVProtector) Unprotect(path string) error {
	return p.setProtected(path, "false")
}

// IsProtected returns true if the secret at path has protected=true in its metadata.
func (p *KVProtector) IsProtected(path string) (bool, error) {
	url := fmt.Sprintf("/v1/%s/metadata/%s", p.mount, path)
	req := p.client.NewRequest(http.MethodGet, url)
	resp, err := p.client.RawRequest(req)
	if err != nil {
		return false, fmt.Errorf("metadata GET failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return false, fmt.Errorf("parse error: %w", err)
	}
	cm, ok := secret.Data["custom_metadata"].(map[string]interface{})
	if !ok {
		return false, nil
	}
	return cm["protected"] == "true", nil
}

// ToggleProtection flips the protection flag on the given path.
// If the secret is currently protected it will be unprotected, and vice versa.
// It returns the new protection state.
func (p *KVProtector) ToggleProtection(path string) (bool, error) {
	protected, err := p.IsProtected(path)
	if err != nil {
		return false, fmt.Errorf("toggle: failed to read current state: %w", err)
	}
	if protected {
		return false, p.Unprotect(path)
	}
	return true, p.Protect(path)
}

func (p *KVProtector) setProtected(path, value string) error {
	url := fmt.Sprintf("/v1/%s/metadata/%s", p.mount, path)
	body := map[string]interface{}{
		"custom_metadata": map[string]string{"protected": value},
	}
	req := p.client.NewRequest(http.MethodPost, url)
	if err := req.SetJSONBody(body); err != nil {
		return fmt.Errorf("set body: %w", err)
	}
	resp, err := p.client.RawRequest(req)
	if err != nil {
		return fmt.Errorf("metadata POST failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
