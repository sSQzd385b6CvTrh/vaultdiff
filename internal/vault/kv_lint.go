package vault

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/vault/api"
)

// LintResult holds a single lint warning for a secret key.
type LintResult struct {
	Key     string
	Version int
	Warning string
}

// KVLinter inspects secret data for common issues.
type KVLinter struct {
	client *api.Client
	mount  string
}

// NewKVLinter creates a KVLinter with the given client and KV mount.
func NewKVLinter(client *api.Client, mount string) *KVLinter {
	if mount == "" {
		mount = "secret"
	}
	return &KVLinter{client: client, mount: mount}
}

// Lint fetches the latest version of the secret at path and returns lint warnings.
func (l *KVLinter) Lint(path string) ([]LintResult, error) {
	url := fmt.Sprintf("/v1/%s/data/%s", l.mount, path)
	resp, err := l.client.RawRequest(l.client.NewRequest(http.MethodGet, url))
	if err != nil {
		return nil, fmt.Errorf("lint fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, path)
	}

	var result struct {
		Data struct {
			Data     map[string]interface{} `json:"data"`
			Metadata struct {
				Version int `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	version := result.Data.Metadata.Version
	var warnings []LintResult
	for k, v := range result.Data.Data {
		warnings = append(warnings, lintField(k, fmt.Sprintf("%v", v), version)...)
	}
	return warnings, nil
}

func lintField(key, value string, version int) []LintResult {
	var results []LintResult
	if value == "" {
		results = append(results, LintResult{Key: key, Version: version, Warning: "empty value"})
	}
	if strings.ToLower(key) != key {
		results = append(results, LintResult{Key: key, Version: version, Warning: "key contains uppercase letters"})
	}
	if strings.Contains(value, " ") && !strings.Contains(strings.ToLower(key), "description") {
		results = append(results, LintResult{Key: key, Version: version, Warning: "value contains spaces"})
	}
	return results
}
