package vault

import (
	"context"
	"fmt"
)

// EnvComparer compares secrets across two named environments.
type EnvComparer struct {
	envs    map[string]*Fetcher
	compare *Comparer
}

// EnvConfig holds the address and optional namespace for a single environment.
type EnvConfig struct {
	Address   string
	Namespace string
	Token     string
}

// NewEnvComparer creates an EnvComparer for the given named environments.
func NewEnvComparer(envs map[string]EnvConfig) (*EnvComparer, error) {
	fetchers := make(map[string]*Fetcher, len(envs))
	for name, cfg := range envs {
		client, err := NewClient(cfg.Address, cfg.Token, cfg.Namespace)
		if err != nil {
			return nil, fmt.Errorf("env %q: %w", name, err)
		}
		fetchers[name] = NewFetcher(client)
	}
	return &EnvComparer{
		envs:    fetchers,
		compare: NewComparer(nil), // comparer is stateless; fetchers used directly
	}, nil
}

// Compare fetches the given secret path+version from two named environments
// and returns the raw data maps for each, ready for diffing.
func (e *EnvComparer) Compare(
	ctx context.Context,
	srcEnv, dstEnv, path string,
	srcVersion, dstVersion int,
) (map[string]interface{}, map[string]interface{}, error) {
	srcFetcher, ok := e.envs[srcEnv]
	if !ok {
		return nil, nil, fmt.Errorf("unknown environment: %q", srcEnv)
	}
	dstFetcher, ok := e.envs[dstEnv]
	if !ok {
		return nil, nil, fmt.Errorf("unknown environment: %q", dstEnv)
	}

	srcSecret, err := srcFetcher.GetSecretVersion(ctx, path, srcVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch src (%s@%d): %w", path, srcVersion, err)
	}

	dstSecret, err := dstFetcher.GetSecretVersion(ctx, path, dstVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch dst (%s@%d): %w", path, dstVersion, err)
	}

	return srcSecret.Data, dstSecret.Data, nil
}
