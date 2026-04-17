package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// KVWriter writes a secret to a KV v2 mount.
type KVWriter struct {
	address string
	token   string
	mount   string
	client  *http.Client
}

// NewKVWriter creates a new KVWriter.
func NewKVWriter(address, token, mount string) *KVWriter {
	if mount == "" {
		mount = "secret"
	}
	return &KVWriter{
		address: address,
		token:   token,
		mount:   mount,
		client:  &http.Client{},
	}
}

// Put writes data to the given secret path and returns the new version.
func (w *KVWriter) Put(path string, data map[string]string) (int, error) {
	payload := map[string]interface{}{"data": data}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s/data/%s", w.address, w.mount, path)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", w.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Data struct {
			Version int `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}
	return result.Data.Version, nil
}
