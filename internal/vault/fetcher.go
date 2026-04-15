package vault

import (
	"context"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretVersion holds the data for a specific version of a KV secret.
type SecretVersion struct {
	Path    string
	Version int
	Data    map[string]string
}

// Fetcher retrieves secret versions from Vault KV v2.
type Fetcher struct {
	client *vaultapi.Client
}

// NewFetcher creates a Fetcher backed by the given Vault client.
func NewFetcher(c *vaultapi.Client) *Fetcher {
	return &Fetcher{client: c}
}

// GetSecretVersion fetches a specific version of a KV v2 secret.
// Pass version 0 to retrieve the latest version.
func (f *Fetcher) GetSecretVersion(ctx context.Context, mount, path string, version int) (*SecretVersion, error) {
	kvPath := fmt.Sprintf("%s/data/%s", mount, path)

	params := map[string][]string{}
	if version > 0 {
		params["version"] = []string{fmt.Sprintf("%d", version)}
	}

	secret, err := f.client.Logical().ReadWithDataWithContext(ctx, kvPath, params)
	if err != nil {
		return nil, fmt.Errorf("reading secret %q: %w", kvPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path %q", kvPath)
	}

	rawData, ok := secret.Data["data"]
	if !ok {
		return nil, fmt.Errorf("unexpected KV v2 response: missing 'data' key")
	}

	dataMap, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected KV v2 response: 'data' is not a map")
	}

	result := &SecretVersion{
		Path:    path,
		Version: version,
		Data:    make(map[string]string, len(dataMap)),
	}

	for k, v := range dataMap {
		result.Data[k] = fmt.Sprintf("%v", v)
	}

	return result, nil
}
