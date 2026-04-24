package vault

import (
	"fmt"
	"sort"
)

// KVEnvDiffResult holds the diff result between two environments for a single key.
type KVEnvDiffResult struct {
	Key    string
	Left   map[string]string
	Right  map[string]string
	Status string // "added", "removed", "modified", "unchanged"
}

// KVEnvDiffer compares KV secrets between two named environments.
type KVEnvDiffer struct {
	left  *KVGetter
	right *KVGetter
}

// NewKVEnvDiffer creates a KVEnvDiffer using two pre-configured KVGetter instances.
func NewKVEnvDiffer(left, right *KVGetter) *KVEnvDiffer {
	return &KVEnvDiffer{left: left, right: right}
}

// Compare fetches the secret at path from both environments and returns a diff result.
func (d *KVEnvDiffer) Compare(mount, path string) (*KVEnvDiffResult, error) {
	lData, lErr := d.left.Get(mount, path)
	rData, rErr := d.right.Get(mount, path)

	lMissing := lErr != nil
	rMissing := rErr != nil

	if lMissing && rMissing {
		return nil, fmt.Errorf("secret %q not found in either environment", path)
	}

	result := &KVEnvDiffResult{
		Key:   path,
		Left:  lData,
		Right: rData,
	}

	switch {
	case lMissing:
		result.Status = "added"
		result.Left = map[string]string{}
	case rMissing:
		result.Status = "removed"
		result.Right = map[string]string{}
	case mapsEqual(lData, rData):
		result.Status = "unchanged"
	default:
		result.Status = "modified"
	}

	return result, nil
}

// SortedKeys returns a sorted union of keys from both data maps.
func (r *KVEnvDiffResult) SortedKeys() []string {
	seen := map[string]struct{}{}
	for k := range r.Left {
		seen[k] = struct{}{}
	}
	for k := range r.Right {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
