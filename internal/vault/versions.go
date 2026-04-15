package vault

import (
	"context"
	"fmt"
	"sort"
)

// VersionMeta holds metadata about a single secret version.
type VersionMeta struct {
	Version      int
	CreatedTime  string
	DeletionTime string
	Destroyed    bool
}

// VersionLister lists available versions for a KV v2 secret path.
type VersionLister struct {
	fetcher *Fetcher
}

// NewVersionLister creates a VersionLister backed by the given Fetcher.
func NewVersionLister(f *Fetcher) *VersionLister {
	return &VersionLister{fetcher: f}
}

// ListVersions returns metadata for all versions of the secret at path.
// path should be the logical secret path (e.g. "secret/myapp/config").
func (vl *VersionLister) ListVersions(ctx context.Context, mount, path string) ([]VersionMeta, error) {
	fullPath := fmt.Sprintf("%s/metadata/%s", mount, path)

	secret, err := vl.fetcher.client.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return nil, fmt.Errorf("listing versions for %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no metadata found for path: %s", path)
	}

	versionsRaw, ok := secret.Data["versions"]
	if !ok {
		return nil, fmt.Errorf("metadata response missing 'versions' key for path: %s", path)
	}

	versionsMap, ok := versionsRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for versions data")
	}

	var metas []VersionMeta
	for _, v := range versionsMap {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		meta := VersionMeta{}
		if ct, ok := vm["created_time"].(string); ok {
			meta.CreatedTime = ct
		}
		if dt, ok := vm["deletion_time"].(string); ok {
			meta.DeletionTime = dt
		}
		if d, ok := vm["destroyed"].(bool); ok {
			meta.Destroyed = d
		}
		metas = append(metas, meta)
	}

	sort.Slice(metas, func(i, j int) bool {
		return metas[i].Version < metas[j].Version
	})

	return metas, nil
}
