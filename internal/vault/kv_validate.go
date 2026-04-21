package vault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// ValidationResult holds the outcome of a secret validation check.
type ValidationResult struct {
	Path    string
	Version int
	Valid   bool
	Reason  string
}

// KVValidator checks whether a secret at a given path and version is
// readable and non-destroyed/non-deleted.
type KVValidator struct {
	client *api.Client
	mount  string
}

// NewKVValidator returns a KVValidator. mount defaults to "secret" if empty.
func NewKVValidator(client *api.Client, mount string) *KVValidator {
	if mount == "" {
		mount = "secret"
	}
	return &KVValidator{client: client, mount: mount}
}

// Validate checks the given path at the specified version.
// A version of 0 means the latest version.
func (v *KVValidator) Validate(path string, version int) (*ValidationResult, error) {
	url := fmt.Sprintf("/v1/%s/data/%s", v.mount, path)
	if version > 0 {
		url = fmt.Sprintf("%s?version=%d", url, version)
	}

	req := v.client.NewRequest(http.MethodGet, url)
	resp, err := v.client.RawRequest(req)
	if err != nil {
		return nil, fmt.Errorf("validate request failed: %w", err)
	}
	defer resp.Body.Close()

	result := &ValidationResult{Path: path, Version: version}

	if resp.StatusCode == http.StatusNotFound {
		result.Valid = false
		result.Reason = "secret not found"
		return result, nil
	}
	if resp.StatusCode != http.StatusOK {
		result.Valid = false
		result.Reason = fmt.Sprintf("unexpected status %d", resp.StatusCode)
		return result, nil
	}

	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	meta, ok := secret.Data["metadata"].(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Reason = "missing metadata"
		return result, nil
	}

	if destroyed, _ := meta["destroyed"].(bool); destroyed {
		result.Valid = false
		result.Reason = "version is destroyed"
		return result, nil
	}
	if deletionTime, _ := meta["deletion_time"].(string); deletionTime != "" {
		result.Valid = false
		result.Reason = fmt.Sprintf("version deleted at %s", deletionTime)
		return result, nil
	}

	result.Valid = true
	return result, nil
}
