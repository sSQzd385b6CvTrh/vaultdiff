package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Rollbacker restores a KV v2 secret to a previous version by re-writing its data.
type Rollbacker struct {
	client *api.Client
	mount  string
}

// NewRollbacker creates a new Rollbacker for the given KV mount.
func NewRollbacker(client *api.Client, mount string) *Rollbacker {
	if mount == "" {
		mount = "secret"
	}
	return &Rollbacker{client: client, mount: mount}
}

// RollbackResult describes the outcome of a rollback.
type RollbackResult struct {
	Path        string
	ToVersion   int
	NewVersion  int
}

// Rollback reads the data at toVersion and writes it as a new version,
// effectively rolling the secret back.
func (r *Rollbacker) Rollback(ctx context.Context, path string, toVersion int) (*RollbackResult, error) {
	readPath := fmt.Sprintf("%s/data/%s?version=%d", r.mount, path, toVersion)
	secret, err := r.client.Logical().ReadWithContext(ctx, readPath)
	if err != nil {
		return nil, fmt.Errorf("rollback read %q v%d: %w", path, toVersion, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("rollback: version %d of %q not found or destroyed", toVersion, path)
	}

	kvData, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("rollback: malformed data for %q", path)
	}

	writePath := fmt.Sprintf("%s/data/%s", r.mount, path)
	result, err := r.client.Logical().WriteWithContext(ctx, writePath, map[string]interface{}{
		"data": kvData,
	})
	if err != nil {
		return nil, fmt.Errorf("rollback write %q: %w", path, err)
	}

	var newVersion int
	if result != nil && result.Data != nil {
		if meta, ok := result.Data["version"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", meta), "%d", &newVersion)
		}
	}

	return &RollbackResult{
		Path:       path,
		ToVersion:  toVersion,
		NewVersion: newVersion,
	}, nil
}
