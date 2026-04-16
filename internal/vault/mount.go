package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

// MountInfo holds basic metadata about a Vault secret engine mount.
type MountInfo struct {
	Path        string
	Type        string
	Description string
	Local       bool
}

// MountLister lists secret engine mounts from Vault.
type MountLister struct {
	client *http.Client
	baseURL string
	token   string
}

// NewMountLister creates a MountLister using the provided Vault client config.
func NewMountLister(address, token string) *MountLister {
	return &MountLister{
		client:  &http.Client{},
		baseURL: address,
		token:   token,
	}
}

// ListMounts returns all secret engine mounts sorted by path.
func (m *MountLister) ListMounts() ([]MountInfo, error) {
	req, err := http.NewRequest(http.MethodGet, m.baseURL+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", m.token)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied listing mounts")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	var raw map[string]struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		Local       bool   `json:"local"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	mounts := make([]MountInfo, 0, len(raw))
	for path, info := range raw {
		mounts = append(mounts, MountInfo{
			Path:        path,
			Type:        info.Type,
			Description: info.Description,
			Local:       info.Local,
		})
	}
	sort.Slice(mounts, func(i, j int) bool { return mounts[i].Path < mounts[j].Path })
	return mounts, nil
}
