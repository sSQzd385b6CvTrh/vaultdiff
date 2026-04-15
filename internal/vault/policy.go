package vault

import (
	"context"
	"fmt"
	"strings"
)

// PolicyLister lists Vault policies for a given path prefix.
type PolicyLister struct {
	client *Client
}

// PolicyEntry represents a single Vault policy with its name and rules.
type PolicyEntry struct {
	Name  string
	Rules string
}

// NewPolicyLister creates a new PolicyLister using the provided Client.
func NewPolicyLister(c *Client) *PolicyLister {
	return &PolicyLister{client: c}
}

// ListPolicies returns all policy names from Vault.
func (p *PolicyLister) ListPolicies(ctx context.Context) ([]string, error) {
	path := "sys/policies/acl"
	secret, err := p.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("listing policies at %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}
	raw, ok := secret.Data["keys"]
	if !ok {
		return []string{}, nil
	}
	ifaces, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for policy keys")
	}
	names := make([]string, 0, len(ifaces))
	for _, v := range ifaces {
		if s, ok := v.(string); ok {
			names = append(names, s)
		}
	}
	return names, nil
}

// GetPolicy retrieves the rules for a named policy.
func (p *PolicyLister) GetPolicy(ctx context.Context, name string) (*PolicyEntry, error) {
	path := fmt.Sprintf("sys/policies/acl/%s", strings.TrimPrefix(name, "/"))
	secret, err := p.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading policy %s: %w", name, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("policy %s not found", name)
	}
	rules, _ := secret.Data["policy"].(string)
	return &PolicyEntry{Name: name, Rules: rules}, nil
}
