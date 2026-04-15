package vault

import (
	"context"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// SealStatus holds the seal state of a Vault instance.
type SealStatus struct {
	Sealed      bool
	Initialized bool
	Progress    int
	Threshold   int
	Total       int
	Version     string
}

// SealChecker checks the seal status of a Vault instance.
type SealChecker struct {
	client *vaultapi.Client
}

// NewSealChecker creates a new SealChecker using the provided Vault client.
func NewSealChecker(client *vaultapi.Client) *SealChecker {
	return &SealChecker{client: client}
}

// Status returns the current seal status of the Vault instance.
func (s *SealChecker) Status(ctx context.Context) (*SealStatus, error) {
	req := s.client.NewRequest(http.MethodGet, "/v1/sys/seal-status")
	resp, err := s.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("seal status request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Sealed      bool   `json:"sealed"`
		Initialized bool   `json:"initialized"`
		Progress    int    `json:"progress"`
		T           int    `json:"t"`
		N           int    `json:"n"`
		Version     string `json:"version"`
	}

	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("failed to decode seal status response: %w", err)
	}

	return &SealStatus{
		Sealed:      result.Sealed,
		Initialized: result.Initialized,
		Progress:    result.Progress,
		Threshold:   result.T,
		Total:       result.N,
		Version:     result.Version,
	}, nil
}
