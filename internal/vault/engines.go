package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

// SecretEngine represents a mounted secret engine in Vault.
type SecretEngine struct {
	Path        string
	Type        string
	Description string
	Accessor    string
}

// EngineListResult holds a list of secret engines.
type EngineListResult struct {
	Engines []SecretEngine
}

// SecretEngineLister lists mounted secret engines.
type SecretEngineLister struct {
	client *Client
}

// NewSecretEngineLister creates a new SecretEngineLister.
func NewSecretEngineLister(c *Client) *SecretEngineLister {
	return &SecretEngineLister{client: c}
}

// ListEngines returns all mounted secret engines from Vault.
func (l *SecretEngineLister) ListEngines() (*EngineListResult, error) {
	req, err := http.NewRequest(http.MethodGet, l.client.Address+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", l.client.Token)
	if l.client.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", l.client.Namespace)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied listing secret engines")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	type mountInfo struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		Accessor    string `json:"accessor"`
	}

	var engines []SecretEngine
	for key, val := range raw {
		var info mountInfo
		if err := json.Unmarshal(val, &info); err != nil {
			continue
		}
		if info.Type == "" {
			continue
		}
		engines = append(engines, SecretEngine{
			Path:        key,
			Type:        info.Type,
			Description: info.Description,
			Accessor:    info.Accessor,
		})
	}
	sort.Slice(engines, func(i, j int) bool { return engines[i].Path < engines[j].Path })
	return &EngineListResult{Engines: engines}, nil
}
