package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// AuditTrailEntry represents a single version entry in a secret's audit trail.
type AuditTrailEntry struct {
	Version     int
	CreatedTime time.Time
	DeletedTime *time.Time
	Destroyed   bool
}

// KVAuditTrailer fetches the full version audit trail for a KV secret.
type KVAuditTrailer struct {
	address string
	token   string
	mount   string
	client  *http.Client
}

// NewKVAuditTrailer creates a new KVAuditTrailer.
func NewKVAuditTrailer(address, token, mount string) *KVAuditTrailer {
	if mount == "" {
		mount = "secret"
	}
	return &KVAuditTrailer{
		address: address,
		token:   token,
		mount:   mount,
		client:  &http.Client{},
	}
}

// GetAuditTrail returns all version entries for the given secret path.
func (t *KVAuditTrailer) GetAuditTrail(path string) ([]AuditTrailEntry, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", t.address, t.mount, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", t.token)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
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
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var entries []AuditTrailEntry
	for k, v := range result.Data.Versions {
		var vnum int
		fmt.Sscanf(k, "%d", &vnum)
		entry := AuditTrailEntry{Version: vnum, Destroyed: v.Destroyed}
		if v.CreatedTime != "" {
			entry.CreatedTime, _ = time.Parse(time.RFC3339Nano, v.CreatedTime)
		}
		if v.DeletionTime != "" {
			dt, _ := time.Parse(time.RFC3339Nano, v.DeletionTime)
			entry.DeletedTime = &dt
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Version < entries[j].Version })
	return entries, nil
}
