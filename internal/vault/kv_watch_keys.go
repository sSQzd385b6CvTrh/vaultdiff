package vault

import (
	"context"
	"fmt"
	"net/http"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// KeyWatchResult holds the result of a key watch poll.
type KeyWatchResult struct {
	Path    string
	Version int
	Changed bool
	Data    map[string]interface{}
}

// KVKeyWatcher watches a list of KV paths for version changes.
type KVKeyWatcher struct {
	client    *vaultapi.Client
	mount     string
	interval  time.Duration
}

// NewKVKeyWatcher creates a new KVKeyWatcher.
func NewKVKeyWatcher(client *vaultapi.Client, mount string, interval time.Duration) *KVKeyWatcher {
	if mount == "" {
		mount = "secret"
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}
	return &KVKeyWatcher{client: client, mount: mount, interval: interval}
}

// WatchKeys polls the given paths at the configured interval and sends results
// to the returned channel whenever a version change is detected.
func (w *KVKeyWatcher) WatchKeys(ctx context.Context, paths []string) (<-chan KeyWatchResult, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one path is required")
	}

	versions := make(map[string]int)
	ch := make(chan KeyWatchResult, len(paths))

	go func() {
		defer close(ch)
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, path := range paths {
					result, err := w.poll(path, versions[path])
					if err != nil {
						continue
					}
					if result.Changed {
						versions[path] = result.Version
						ch <- result
					}
				}
			}
		}
	}()

	return ch, nil
}

func (w *KVKeyWatcher) poll(path string, knownVersion int) (KeyWatchResult, error) {
	url := fmt.Sprintf("/v1/%s/data/%s", w.mount, path)
	req := w.client.NewRequest(http.MethodGet, url)
	resp, err := w.client.RawRequest(req)
	if err != nil {
		return KeyWatchResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return KeyWatchResult{}, fmt.Errorf("unexpected status %d for path %s", resp.StatusCode, path)
	}
	var body struct {
		Data struct {
			Data     map[string]interface{} `json:"data"`
			Metadata struct {
				Version int `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := resp.DecodeJSON(&body); err != nil {
		return KeyWatchResult{}, err
	}
	v := body.Data.Metadata.Version
	return KeyWatchResult{
		Path:    path,
		Version: v,
		Changed: v != knownVersion,
		Data:    body.Data.Data,
	}, nil
}
