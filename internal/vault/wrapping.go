package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WrappingInfo holds details about a wrapped token response.
type WrappingInfo struct {
	Token          string `json:"token"`
	Accessor       string `json:"accessor"`
	TTL            int    `json:"ttl"`
	CreationTime   string `json:"creation_time"`
	CreationPath   string `json:"creation_path"`
	WrappedAccessor string `json:"wrapped_accessor"`
}

// WrappingInspector looks up wrapping token metadata.
type WrappingInspector struct {
	client *Client
}

// NewWrappingInspector returns a new WrappingInspector.
func NewWrappingInspector(c *Client) *WrappingInspector {
	return &WrappingInspector{client: c}
}

// Lookup calls sys/wrapping/lookup with the given token.
func (w *WrappingInspector) Lookup(token string) (*WrappingInfo, error) {
	body := map[string]string{"token": token}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := newJSONRequest(w.client.Address, "POST", "/v1/sys/wrapping/lookup", b)
	if err != nil {
		return nil, err
	}
	if w.client.Token != "" {
		req.Header.Set("X-Vault-Token", w.client.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("wrapping token not found or already unwrapped")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result struct {
		Data WrappingInfo `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &result.Data, nil
}
