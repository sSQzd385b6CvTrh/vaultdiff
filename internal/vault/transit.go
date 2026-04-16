package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TransitKey holds metadata about a Vault transit encryption key.
type TransitKey struct {
	Name            string
	Type            string
	DeletionAllowed bool
	Exportable      bool
	LatestVersion   int
}

// TransitLister lists transit keys from Vault.
type TransitLister struct {
	Address string
	Token   string
	Mount   string
}

// NewTransitLister creates a TransitLister with the given config.
func NewTransitLister(address, token, mount string) *TransitLister {
	if mount == "" {
		mount = "transit"
	}
	return &TransitLister{Address: address, Token: token, Mount: mount}
}

// ListKeys returns all transit key names under the configured mount.
func (t *TransitLister) ListKeys() ([]string, error) {
	url := fmt.Sprintf("%s/v1/%s/keys?list=true", t.Address, t.Mount)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", t.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data.Keys, nil
}
