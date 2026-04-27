package vault

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/hashicorp/vault/api"
)

// AccessLogEntry represents a single version access record derived from KV metadata.
type AccessLogEntry struct {
	Version      int
	CreatedTime  string
	DeletionTime string
	Destroyed    bool
}

// KVAccessLogger retrieves version-level access metadata for a KV secret.
type KVAccessLogger struct {
	client *api.Client
	mount  string
}

// NewKVAccessLogger creates a new KVAccessLogger.
func NewKVAccessLogger(client *api.Client, mount string) *KVAccessLogger {
	if mount == "" {
		mount = "secret"
	}
	return &KVAccessLogger{client: client, mount: mount}
}

// GetAccessLog returns a sorted list of AccessLogEntry for the given secret path.
func (l *KVAccessLogger) GetAccessLog(path string) ([]AccessLogEntry, error) {
	url := fmt.Sprintf("/v1/%s/metadata/%s", l.mount, path)
	req := l.client.NewRequest(http.MethodGet, url)

	resp, err := l.client.RawRequest(req)
	if err != nil {
		return nil, fmt.Errorf("access log request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Versions map[string]struct {
				CreatedTime  string `json:"created_time"`
				DeletionTime string `json:"deletion_time"`
				Destroyed    bool   `json:"destroyed"`
			} `json:"versions"`
		} `json:"data"`
	}

	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	var entries []AccessLogEntry
	for vStr, meta := range result.Data.Versions {
		var vNum int
		fmt.Sscanf(vStr, "%d", &vNum)
		entries = append(entries, AccessLogEntry{
			Version:      vNum,
			CreatedTime:  meta.CreatedTime,
			DeletionTime: meta.DeletionTime,
			Destroyed:    meta.Destroyed,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Version < entries[j].Version
	})

	_ = time.Now() // ensure time import used
	return entries, nil
}
