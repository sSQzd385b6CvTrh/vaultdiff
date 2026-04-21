package vault

import (
	"fmt"
	"net/http"
	"time"
)

// LockInfo holds the lock state for a KV secret path.
type LockInfo struct {
	Path      string
	Locked    bool
	LockedBy  string
	LockedAt  time.Time
	ExpiresAt time.Time
}

// KVLocker manages advisory locks on KV secret paths via metadata custom_metadata.
type KVLocker struct {
	client *Client
	mount  string
}

// NewKVLocker creates a new KVLocker. Defaults to "secret" mount if empty.
func NewKVLocker(client *Client, mount string) *KVLocker {
	if mount == "" {
		mount = "secret"
	}
	return &KVLocker{client: client, mount: mount}
}

// Lock sets an advisory lock on the given path using KV metadata custom_metadata.
func (l *KVLocker) Lock(path, owner string, ttl time.Duration) error {
	now := time.Now().UTC()
	expires := now.Add(ttl)

	body := map[string]interface{}{
		"custom_metadata": map[string]string{
			"locked":     "true",
			"locked_by":  owner,
			"locked_at":  now.Format(time.RFC3339),
			"expires_at": expires.Format(time.RFC3339),
		},
	}

	url := fmt.Sprintf("%s/v1/%s/metadata/%s", l.client.Address, l.mount, path)
	resp, err := l.client.RawRequest(http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("lock request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status locking %s: %d", path, resp.StatusCode)
	}
	return nil
}

// Unlock removes the advisory lock from the given path.
func (l *KVLocker) Unlock(path string) error {
	body := map[string]interface{}{
		"custom_metadata": map[string]string{
			"locked":     "false",
			"locked_by":  "",
			"locked_at":  "",
			"expires_at": "",
		},
	}

	url := fmt.Sprintf("%s/v1/%s/metadata/%s", l.client.Address, l.mount, path)
	resp, err := l.client.RawRequest(http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("unlock request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status unlocking %s: %d", path, resp.StatusCode)
	}
	return nil
}
