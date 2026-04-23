package vault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// SnapshotDiffResult holds the comparison between two KV snapshots.
type SnapshotDiffResult struct {
	Added    map[string]string
	Removed  map[string]string
	Modified map[string]string
	Unchanged map[string]string
}

// KVSnapshotDiffer compares two KV secret snapshots by path.
type KVSnapshotDiffer struct {
	client *api.Client
	mount  string
}

// NewKVSnapshotDiffer creates a new KVSnapshotDiffer.
func NewKVSnapshotDiffer(client *api.Client, mount string) *KVSnapshotDiffer {
	if mount == "" {
		mount = "secret"
	}
	return &KVSnapshotDiffer{client: client, mount: mount}
}

// Compare fetches both paths and returns a SnapshotDiffResult.
func (d *KVSnapshotDiffer) Compare(pathA, pathB string) (*SnapshotDiffResult, error) {
	aData, err := d.fetch(pathA)
	if err != nil {
		return nil, fmt.Errorf("fetch %q: %w", pathA, err)
	}
	bData, err := d.fetch(pathB)
	if err != nil {
		return nil, fmt.Errorf("fetch %q: %w", pathB, err)
	}

	result := &SnapshotDiffResult{
		Added:    make(map[string]string),
		Removed:  make(map[string]string),
		Modified: make(map[string]string),
		Unchanged: make(map[string]string),
	}

	for k, v := range bData {
		if orig, ok := aData[k]; !ok {
			result.Added[k] = v
		} else if orig != v {
			result.Modified[k] = v
		} else {
			result.Unchanged[k] = v
		}
	}
	for k, v := range aData {
		if _, ok := bData[k]; !ok {
			result.Removed[k] = v
		}
	}
	return result, nil
}

func (d *KVSnapshotDiffer) fetch(path string) (map[string]string, error) {
	url := fmt.Sprintf("/v1/%s/data/%s", d.mount, path)
	req := d.client.NewRequest(http.MethodGet, url)
	resp, err := d.client.RawRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var body struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := resp.DecodeJSON(&body); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(body.Data.Data))
	for k, v := range body.Data.Data {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out, nil
}
