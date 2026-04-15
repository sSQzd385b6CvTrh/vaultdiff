package vault

import (
	"context"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods for secret versioning.
type Client struct {
	api *vaultapi.Client
}

// Config holds connection configuration for a Vault instance.
type Config struct {
	Address   string
	Token     string
	Namespace string
}

// SecretVersion represents a single version of a KV v2 secret.
type SecretVersion struct {
	Version  int
	Data     map[string]string
	Metadata map[string]interface{}
}

// NewClient creates a new Vault client from the provided config.
func NewClient(cfg Config) (*Client, error) {
	vCfg := vaultapi.DefaultConfig()
	if cfg.Address != "" {
		vCfg.Address = cfg.Address
	}

	client, err := vaultapi.NewClient(vCfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	if cfg.Token != "" {
		client.SetToken(cfg.Token)
	}

	if cfg.Namespace != "" {
		client.SetNamespace(cfg.Namespace)
	}

	return &Client{api: client}, nil
}

// GetSecretVersion fetches a specific version of a KV v2 secret.
// Pass version 0 to retrieve the latest version.
func (c *Client) GetSecretVersion(ctx context.Context, mount, path string, version int) (*SecretVersion, error) {
	params := map[string][]string{}
	if version > 0 {
		params["version"] = []string{fmt.Sprintf("%d", version)}
	}

	secret, err := c.api.KVv2(mount).GetVersion(ctx, path, version)
	if err != nil {
		return nil, fmt.Errorf("fetching secret %s/%s@v%d: %w", mount, path, version, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret %s/%s not found", mount, path)
	}

	data := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		data[k] = fmt.Sprintf("%v", v)
	}

	return &SecretVersion{
		Version:  version,
		Data:     data,
		Metadata: map[string]interface{}{"created_time": secret.VersionMetadata.CreatedTime},
	}, nil
}
