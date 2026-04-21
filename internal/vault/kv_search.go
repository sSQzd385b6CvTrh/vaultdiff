package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// KVSearchResult holds a matched secret path and the matching keys.
type KVSearchResult struct {
	Path string
	MatchedKeys []string
}

// KVSearcher searches for secrets whose keys or values match a query.
type KVSearcher struct {
	client *api.Client
	mount  string
}

// NewKVSearcher creates a new KVSearcher.
func NewKVSearcher(client *api.Client, mount string) *KVSearcher {
	if mount == "" {
		mount = "secret"
	}
	return &KVSearcher{client: client, mount: mount}
}

// Search lists all secrets under prefix and returns those whose keys match query.
func (s *KVSearcher) Search(ctx context.Context, prefix, query string) ([]KVSearchResult, error) {
	paths, err := s.listAll(ctx, prefix)
	if err != nil {
		return nil, err
	}

	var results []KVSearchResult
	for _, path := range paths {
		secret, err := s.client.Logical().ReadWithContext(ctx,
			fmt.Sprintf("%s/data/%s", s.mount, path))
		if err != nil || secret == nil ||			continue
		}
		data, ok := secret.Data["data"].(map[string]interface{})
		if !ok {
			continue
		}
		var matched []string
		for k := range data {
			if strings.Contains(k, query) {
				matched = append(matched, k)
			}
		}
		if len(matched) > 0 {
			results = append(results, KVSearchResult{Path: path, MatchedKeys: matched})
		}
	}
	return results, nil
}

func (s *KVSearcher) listAll(ctx context.Context, prefix string) ([]string, error) {
	secret, err := s.client.Logical().ListWithContext(ctx,
		fmt.Sprintf("%s/metadata/%s", s.mount, prefix))
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", prefix, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}
	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}
	var paths []string
	for _, k := range keys {
		name, _ := k.(string)
		full := strings.TrimSuffix(prefix, "/") + "/" + name
		if strings.HasSuffix(name, "/") {
			sub, err := s.listAll(ctx, full)
			if err != nil {
				return nil, err
			}
			paths = append(paths, sub...)
		} else {
			paths = append(paths, strings.TrimPrefix(full, "/"))
		}
	}
	return paths, nil
}
