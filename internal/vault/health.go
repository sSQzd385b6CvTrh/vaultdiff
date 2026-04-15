package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HealthStatus represents the health check response from Vault.
type HealthStatus struct {
	Initialized bool   `json:"initialized"`
	Sealed      bool   `json:"sealed"`
	Standby     bool   `json:"standby"`
	Version     string `json:"version"`
	ClusterName string `json:"cluster_name"`
	ClusterID   string `json:"cluster_id"`
}

// HealthChecker checks the health of a Vault instance.
type HealthChecker struct {
	address    string
	httpClient *http.Client
}

// NewHealthChecker creates a new HealthChecker for the given Vault address.
func NewHealthChecker(address string) *HealthChecker {
	return &HealthChecker{
		address: address,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Check performs a health check against the Vault instance.
func (h *HealthChecker) Check(ctx context.Context) (*HealthStatus, error) {
	url := fmt.Sprintf("%s/v1/sys/health?standbyok=true&perfstandbyok=true", h.address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building health request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing health request: %w", err)
	}
	defer resp.Body.Close()

	var status HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding health response: %w", err)
	}

	return &status, nil
}
