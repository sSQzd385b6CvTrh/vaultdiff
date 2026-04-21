package vault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// KVDiffResult holds the comparison between two KV secret paths.
type KVDiffResult struct {
	SourcePath  string
	TargetPath  string
	SourceData  map[string]interface{}
	TargetData  map[string]interface{}
}

// KVDiffer fetches and compares two KV secret paths.
type KVDiffer struct {
	client *api.Client
	mount  string
}

// NewKVDiffer creates a new KVDiffer with the given client and mount.
func NewKVDiffer(client *api.Client, mount string) *KVDiffer {
	if mount == "" {
		mount = "secret"
	}
	return &KVDiffer{client: client, mount: mount}
}

// Diff fetches both secret paths and returns a KVDiffResult.
func (d *KVDiffer) Diff(srcPath, dstPath string) (*KVDiffResult, error) {
	srcData, err := d.fetchSecret(srcPath)
	if err != nil {
		return nil, fmt.Errorf("source %q: %w", srcPath, err)
	}

	dstData, err := d.fetchSecret(dstPath)
	if err != nil {
		return nil, fmt.Errorf("target %q: %w", dstPath, err)
	}

	return &KVDiffResult{
		SourcePath: srcPath,
		TargetPath: dstPath,
		SourceData: srcData,
		TargetData: dstData,
	}, nil
}

func (d *KVDiffer) fetchSecret(path string) (map[string]interface{}, error) {
	url := fmt.Sprintf("/v1/%s/data/%s", d.mount, path)
	resp, err := d.client.RawRequest(d.client.NewRequest(http.MethodGet, url))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	return result.Data.Data, nil
}
