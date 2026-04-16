package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TokenInfo holds metadata about the current Vault token.
type TokenInfo struct {
	Accessor   string
	DisplayName string
	Policies   []string
	TTL        time.Duration
	Renewable  bool
	ExpireTime *time.Time
}

// TokenInspector looks up information about the current Vault token.
type TokenInspector struct {
	client *http.Client
	baseURL string
	token   string
}

// NewTokenInspector creates a TokenInspector using the provided Vault client config.
func NewTokenInspector(address, token string) *TokenInspector {
	return &TokenInspector{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: address,
		token:   token,
	}
}

// LookupSelf queries the /auth/token/lookup-self endpoint.
func (t *TokenInspector) LookupSelf() (*TokenInfo, error) {
	url := fmt.Sprintf("%s/v1/auth/token/lookup-self", t.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", t.token)

	resp, err := t.client.Do(req)
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
		return nil, fmt.Errorf("reading body: %w", err)
	}

	var result struct {
		Data struct {
			Accessor    string   `json:"accessor"`
			DisplayName string   `json:"display_name"`
			Policies    []string `json:"policies"`
			TTL         int      `json:"ttl"`
			Renewable   bool     `json:"renewable"`
			ExpireTime  string   `json:"expire_time"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	info := &TokenInfo{
		Accessor:    result.Data.Accessor,
		DisplayName: result.Data.DisplayName,
		Policies:    result.Data.Policies,
		TTL:         time.Duration(result.Data.TTL) * time.Second,
		Renewable:   result.Data.Renewable,
	}

	if result.Data.ExpireTime != "" {
		parsed, err := time.Parse(time.RFC3339, result.Data.ExpireTime)
		if err == nil {
			info.ExpireTime = &parsed
		}
	}

	return info, nil
}
