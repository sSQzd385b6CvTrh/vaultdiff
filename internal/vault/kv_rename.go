package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// KVRenamer renames a KV secret by copying it to a new path and deleting the original.
type KVRenamer struct {
	client *api.Client
	mount  string
}

// NewKVRenamer returns a KVRenamer for the given mount.
func NewKVRenamer(client *api.Client, mount string) *KVRenamer {
	if mount == "" {
		mount = "secret"
	}
	return &KVRenamer{client: client, mount: mount}
}

// Rename copies the secret at srcPath to dstPath, then deletes the source.
// Both paths are relative to the mount point.
func (r *KVRenamer) Rename(ctx context.Context, srcPath, dstPath string) error {
	readPath := fmt.Sprintf("%s/data/%s", r.mount, srcPath)
	secret, err := r.client.Logical().ReadWithContext(ctx, readPath)
	if err != nil {
		return fmt.Errorf("rename: read source %q: %w", srcPath, err)
	}
	if secret == nil || secret.Data == nil {
		return fmt.Errorf("rename: source path %q not found", srcPath)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("rename: unexpected data format at %q", srcPath)
	}

	writePath := fmt.Sprintf("%s/data/%s", r.mount, dstPath)
	_, err = r.client.Logical().WriteWithContext(ctx, writePath, map[string]interface{}{"data": data})
	if err != nil {
		return fmt.Errorf("rename: write destination %q: %w", dstPath, err)
	}

	deletePath := fmt.Sprintf("%s/metadata/%s", r.mount, srcPath)
	_, err = r.client.Logical().DeleteWithContext(ctx, deletePath)
	if err != nil {
		return fmt.Errorf("rename: delete source %q: %w", srcPath, err)
	}

	return nil
}
