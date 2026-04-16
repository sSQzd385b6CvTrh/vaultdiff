package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LeaseInfo holds metadata about a Vault lease.
type LeaseInfo struct {
	LeaseID       string        `json:"lease_id"`
	Renewable     bool          `json:"renewable"`
	LeaseDuration time.Duration `json:"lease_duration"`
}

// LeaseInspector fetches lease information from Vault.
type LeaseInspector struct {
	address    string
	token      string
	httpClient *http.Client
}

// NewLeaseInspector creates a new LeaseInspector.
func NewLeaseInspector(address, token string) *LeaseInspector {
	return &LeaseInspector{
		address:    address,
		token:      token,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Lookup returns lease info for the given lease ID.
func (l *LeaseInspector) Lookup(leaseID string) (*LeaseInfo, error) {
	url := fmt.Sprintf("%s/v1/sys/leases/lookup", l.address)

	body := fmt.Sprintf(`{"lease_id":%q}`, leaseID)
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", l.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("lease not found: %s", leaseID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result struct {
		LeaseID       string `json:"id"`
		Renewable     bool   `json:"renewable"`
		LeaseDuration int    `json:"lease_duration"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &LeaseInfo{
		LeaseID:       result.LeaseID,
		Renewable:     result.Renewable,
		LeaseDuration: time.Duration(result.LeaseDuration) * time.Second,
	}, nil
}
