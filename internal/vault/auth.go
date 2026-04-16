package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AuthInfo holds information about the current token's auth method.
type AuthInfo struct {
	Accessor    string
	DisplayName string
	Policies    []string
	Meta        map[string]string
	TTL         int
	Renewable   bool
}

// AuthInspector fetches auth info for the current token.
type AuthInspector struct {
	address string
	token   string
	client  *http.Client
}

// NewAuthInspector creates a new AuthInspector.
func NewAuthInspector(address, token string) *AuthInspector {
	return &AuthInspector{
		address: address,
		token:   token,
		client:  &http.Client{},
	}
}

// LookupAuth returns auth metadata for the current token.
func (a *AuthInspector) LookupAuth() (*AuthInfo, error) {
	url := fmt.Sprintf("%s/v1/auth/token/lookup-self", a.address)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Vault-Token", a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result struct {
		Data struct {
			Accessor    string            `json:"accessor"`
			DisplayName string            `json:"display_name"`
			Policies    []string          `json:"policies"`
			Meta        map[string]string `json:"meta"`
			TTL         int               `json:"ttl"`
			Renewable   bool              `json:"renewable"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &AuthInfo{
		Accessor:    result.Data.Accessor,
		DisplayName: result.Data.DisplayName,
		Policies:    result.Data.Policies,
		Meta:        result.Data.Meta,
		TTL:         result.Data.TTL,
		Renewable:   result.Data.Renewable,
	}, nil
}
