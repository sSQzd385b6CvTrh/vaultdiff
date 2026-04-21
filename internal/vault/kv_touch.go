package vault

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// KVToucher re-writes a secret at the same path to bump its version.
type KVToucher struct {
	client *api.Client
	mount  string
}

// NewKVToucher creates a KVToucher. If mount is empty it defaults to "secret".
func NewKVToucher(client *api.Client, mount string) *KVToucher {
	if mount == "" {
		mount = "secret"
	}
	return &KVToucher{client: client, mount: mount}
}

// Touch reads the current data at path and writes it back, creating a new version.
func (t *KVToucher) Touch(ctx context.Context, path string) (int, error) {
	// Read current version
	getPath := fmt.Sprintf("/v1/%s/data/%s", t.mount, path)
	req := t.client.NewRequest(http.MethodGet, getPath)
	resp, err := t.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("touch read %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status reading %s: %d", path, resp.StatusCode)
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return 0, fmt.Errorf("decode read response: %w", err)
	}

	// Write back the same data
	putPath := fmt.Sprintf("/v1/%s/data/%s", t.mount, path)
	body := map[string]interface{}{"data": result.Data.Data}
	putReq := t.client.NewRequest(http.MethodPost, putPath)
	if err := putReq.SetJSONBody(body); err != nil {
		return 0, fmt.Errorf("set json body: %w", err)
	}
	putResp, err := t.client.RawRequestWithContext(ctx, putReq)
	if err != nil {
		return 0, fmt.Errorf("touch write %s: %w", path, err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusNoContent {
		return 0, fmt.Errorf("unexpected status writing %s: %d", path, putResp.StatusCode)
	}

	var writeResult struct {
		Data struct {
			Version int `json:"version"`
		} `json:"data"`
	}
	if err := putResp.DecodeJSON(&writeResult); err != nil {
		return 0, fmt.Errorf("decode write response: %w", err)
	}
	return writeResult.Data.Version, nil
}
