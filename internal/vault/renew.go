package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RenewResult holds the result of a token renewal.
type RenewResult struct {
	ClientToken   string
	LeaseDuration time.Duration
	Renewable     bool
}

// TokenRenewer renews Vault tokens.
type TokenRenewer struct {
	address string
	token   string
	client  *http.Client
}

// NewTokenRenewer creates a new TokenRenewer.
func NewTokenRenewer(address, token string) *TokenRenewer {
	return &TokenRenewer{
		address: address,
		token:   token,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Renew renews the current token and returns the result.
func (r *TokenRenewer) Renew() (*RenewResult, error) {
	url := fmt.Sprintf("%s/v1/auth/token/renew-self", r.address)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.token)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("token is not renewable or permission denied")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var payload struct {
		Auth struct {
			ClientToken   string `json:"client_token"`
			LeaseDuration int    `json:"lease_duration"`
			Renewable     bool   `json:"renewable"`
		} `json:"auth"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &RenewResult{
		ClientToken:   payload.Auth.ClientToken,
		LeaseDuration: time.Duration(payload.Auth.LeaseDuration) * time.Second,
		Renewable:     payload.Auth.Renewable,
	}, nil
}
