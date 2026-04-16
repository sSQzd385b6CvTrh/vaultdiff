package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Namespace represents a Vault namespace entry.
type Namespace struct {
	Path string
	ID   string
}

// NamespaceLister lists child namespaces under a given path.
type NamespaceLister struct {
	address string
	token   string
	client  *http.Client
}

// NewNamespaceLister creates a new NamespaceLister.
func NewNamespaceLister(address, token string) *NamespaceLister {
	return &NamespaceLister{
		address: address,
		token:   token,
		client:  &http.Client{},
	}
}

// ListNamespaces returns namespaces under the given parent path.
func (n *NamespaceLister) ListNamespaces(parent string) ([]Namespace, error) {
	url := fmt.Sprintf("%s/v1/sys/namespaces", n.address)
	if parent != "" {
		url = fmt.Sprintf("%s/v1/%s/sys/namespaces", n.address, parent)
	}

	req, err := http.NewRequest("LIST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", n.token)

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("namespace path not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	namespaces := make([]Namespace, 0, len(result.Data.Keys))
	for _, k := range result.Data.Keys {
		namespaces = append(namespaces, Namespace{Path: k})
	}
	return namespaces, nil
}
