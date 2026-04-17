package vault

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// CapabilityChecker checks token capabilities on given paths.
type CapabilityChecker struct {
	client *api.Client
}

// NewCapabilityChecker returns a new CapabilityChecker.
func NewCapabilityChecker(client *api.Client) *CapabilityChecker {
	return &CapabilityChecker{client: client}
}

// CapabilityResult holds the capabilities for a path.
type CapabilityResult struct {
	Path         string
	Capabilities []string
}

// CheckSelf returns the capabilities of the current token on the given paths.
func (c *CapabilityChecker) CheckSelf(ctx context.Context, paths []string) ([]CapabilityResult, error) {
	body := map[string]interface{}{
		"paths": paths,
	}

	req := c.client.NewRequest(http.MethodPost, "/v1/sys/capabilities-self")
	if err := req.SetJSONBody(body); err != nil {
		return nil, fmt.Errorf("set request body: %w", err)
	}

	resp, err := c.client.RawRequestWithContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("capabilities-self request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := resp.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var out []CapabilityResult
	for _, path := range paths {
		caps := []string{}
		if raw, ok := result[path]; ok {
			if list, ok := raw.([]interface{}); ok {
				for _, v := range list {
					if s, ok := v.(string); ok {
						caps = append(caps, s)
					}
				}
			}
		}
		out = append(out, CapabilityResult{Path: path, Capabilities: caps})
	}
	return out, nil
}
