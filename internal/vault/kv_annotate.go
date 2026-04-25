package vault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// KVAnnotation holds a key's custom annotation metadata.
type KVAnnotation struct {
	Path        string
	Annotations map[string]string
}

// KVAnnotator reads and writes custom annotations stored in KV v2 metadata.
type KVAnnotator struct {
	client *api.Client
	mount  string
}

// NewKVAnnotator creates a new KVAnnotator.
func NewKVAnnotator(client *api.Client, mount string) *KVAnnotator {
	if mount == "" {
		mount = "secret"
	}
	return &KVAnnotator{client: client, mount: mount}
}

// GetAnnotations retrieves the custom_metadata annotations for a KV path.
func (a *KVAnnotator) GetAnnotations(path string) (*KVAnnotation, error) {
	url := fmt.Sprintf("/v1/%s/metadata/%s", a.mount, path)
	req := a.client.NewRequest(http.MethodGet, url)
	resp, err := a.client.RawRequest(req)
	if err != nil {
		return nil, fmt.Errorf("get annotations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			CustomMetadata map[string]string `json:"custom_metadata"`
		} `json:"data"`
	}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("decode annotations: %w", err)
	}
	annotations := result.Data.CustomMetadata
	if annotations == nil {
		annotations = map[string]string{}
	}
	return &KVAnnotation{Path: path, Annotations: annotations}, nil
}

// SetAnnotations writes custom_metadata annotations for a KV path.
func (a *KVAnnotator) SetAnnotations(path string, annotations map[string]string) error {
	url := fmt.Sprintf("/v1/%s/metadata/%s", a.mount, path)
	body := map[string]interface{}{
		"custom_metadata": annotations,
	}
	req := a.client.NewRequest(http.MethodPost, url)
	if err := req.SetJSONBody(body); err != nil {
		return fmt.Errorf("set json body: %w", err)
	}
	resp, err := a.client.RawRequest(req)
	if err != nil {
		return fmt.Errorf("set annotations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
