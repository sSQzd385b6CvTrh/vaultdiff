package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// KVCopier copies a secret from one path to another within the same mount.
type KVCopier struct {
	address string
	token   string
	mount   string
	client  *http.Client
}

// NewKVCopier creates a new KVCopier.
func NewKVCopier(address, token, mount string) *KVCopier {
	if mount == "" {
		mount = "secret"
	}
	return &KVCopier{
		address: address,
		token:   token,
		mount:   mount,
		client:  &http.Client{},
	}
}

// Copy reads the secret at srcPath and writes it to dstPath.
func (c *KVCopier) Copy(srcPath, dstPath string) error {
	getURL := fmt.Sprintf("%s/v1/%s/data/%s", c.address, c.mount, srcPath)
	req, err := http.NewRequest(http.MethodGet, getURL, nil)
	if err != nil {
		return fmt.Errorf("build get request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("source secret not found: %s", srcPath)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status reading source: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	writer := NewKVWriter(c.address, c.token, c.mount)
	return writer.Put(dstPath, result.Data.Data)
}
