package vault

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// WatchEvent represents a change detected during secret watching.
type WatchEvent struct {
	Path    string
	Version int
	Data    map[string]interface{}
	Err     error
}

// Watcher polls a Vault KV secret path for version changes.
type Watcher struct {
	client   *vaultapi.Client
	mount    string
	interval time.Duration
}

// NewWatcher creates a Watcher for the given mount and poll interval.
func NewWatcher(client *vaultapi.Client, mount string, interval time.Duration) *Watcher {
	if mount == "" {
		mount = "secret"
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Watcher{client: client, mount: mount, interval: interval}
}

// Watch polls the given secret path and emits events on ch when the version changes.
// It blocks until ctx is cancelled.
func (w *Watcher) Watch(ctx context.Context, path string, ch chan<- WatchEvent) {
	var lastVersion int
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			version, data, err := w.fetchLatest(path)
			if err != nil {
				ch <- WatchEvent{Path: path, Err: err}
				continue
			}
			if version != lastVersion {
				lastVersion = version
				ch <- WatchEvent{Path: path, Version: version, Data: data}
			}
		}
	}
}

func (w *Watcher) fetchLatest(path string) (int, map[string]interface{}, error) {
	apiPath := fmt.Sprintf("%s/data/%s", w.mount, path)
	secret, err := w.client.Logical().Read(apiPath)
	if err != nil {
		return 0, nil, fmt.Errorf("read %s: %w", apiPath, err)
	}
	if secret == nil || secret.Data == nil {
		return 0, nil, fmt.Errorf("path not found: %s", path)
	}
	meta, _ := secret.Data["metadata"].(map[string]interface{})
	version := 0
	if meta != nil {
		if v, ok := meta["version"].(float64); ok {
			version = int(v)
		}
	}
	data, _ := secret.Data["data"].(map[string]interface{})
	return version, data, nil
}
