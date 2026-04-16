package vault

import (
	"fmt"
)

// SecretPair holds two versions of a secret for comparison.
type SecretPair struct {
	Path     string
	VersionA int
	VersionB int
	DataA    map[string]interface{}
	DataB    map[string]interface{}
}

// IsIdentical returns true if both versions contain the same keys and values.
func (sp *SecretPair) IsIdentical() bool {
	if len(sp.DataA) != len(sp.DataB) {
		return false
	}
	for k, v := range sp.DataA {
		if fmt.Sprintf("%v", sp.DataB[k]) != fmt.Sprintf("%v", v) {
			return false
		}
	}
	return true
}

// Comparer fetches and pairs secret versions for diffing.
type Comparer struct {
	fetcher *Fetcher
}

// NewComparer creates a new Comparer backed by the given Fetcher.
func NewComparer(f *Fetcher) *Comparer {
	return &Comparer{fetcher: f}
}

// Compare retrieves two versions of a secret at path and returns a SecretPair.
// If versionA or versionB is 0, the latest version is fetched for that slot.
func (c *Comparer) Compare(mount, path string, versionA, versionB int) (*SecretPair, error) {
	a, err := c.fetcher.GetSecretVersion(mount, path, versionA)
	if err != nil {
		return nil, fmt.Errorf("fetching version %d of %s/%s: %w", versionA, mount, path, err)
	}

	b, err := c.fetcher.GetSecretVersion(mount, path, versionB)
	if err != nil {
		return nil, fmt.Errorf("fetching version %d of %s/%s: %w", versionB, mount, path, err)
	}

	return &SecretPair{
		Path:     fmt.Sprintf("%s/%s", mount, path),
		VersionA: a.Version,
		VersionB: b.Version,
		DataA:    a.Data,
		DataB:    b.Data,
	}, nil
}
