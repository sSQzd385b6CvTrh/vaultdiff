package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RekeyStatus holds the current rekey operation status.
type RekeyStatus struct {
	Started        bool     `json:"started"`
	T              int      `json:"t"`
	N              int      `json:"n"`
	Progress       int      `json:"progress"`
	Required       int      `json:"required"`
	PGPFingerprints []string `json:"pgp_fingerprints"`
	Backup         bool     `json:"backup"`
}

// RekeyChecker checks the status of a Vault rekey operation.
type RekeyChecker struct {
	address string
	token   string
	client  *http.Client
}

// NewRekeyChecker creates a new RekeyChecker.
func NewRekeyChecker(address, token string) *RekeyChecker {
	return &RekeyChecker{
		address: address,
		token:   token,
		client:  &http.Client{},
	}
}

// RekeyStatusResult fetches the current rekey status from Vault.
func (r *RekeyChecker) RekeyStatusResult(ctx context.Context) (*RekeyStatus, error) {
	url := fmt.Sprintf("%s/v1/sys/rekey/init", r.address)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.token)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var status RekeyStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &status, nil
}
