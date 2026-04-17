package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AuditDevice struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Options     map[string]string `json:"options"`
	Path        string
}

type AuditDeviceLister struct {
	client *Client
}

func NewAuditDeviceLister(c *Client) *AuditDeviceLister {
	return &AuditDeviceLister{client: c}
}

func (a *AuditDeviceLister) List() ([]AuditDevice, error) {
	req, err := http.NewRequest(http.MethodGet, a.client.Address+"/v1/sys/audit", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", a.client.Token)
	if a.client.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", a.client.Namespace)
	}

	resp, err := a.client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied")
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

	var devices []AuditDevice
	for path, data := range raw {
		var d AuditDevice
		if err := json.Unmarshal(data, &d); err != nil {
			continue
		}
		d.Path = path
		devices = append(devices, d)
	}
	return devices, nil
}
