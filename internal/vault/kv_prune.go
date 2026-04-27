package vault

import (
	"context"
	"fmt"
	"net/http"
)

// PruneResult holds the outcome of a prune operation on a single key.
type PruneResult struct {
	Key     string
	Pruned  int
	Skipped int
}

// KVPruner deletes all but the N most recent versions of a KV secret.
type KVPruner struct {
	client *Client
	mount  string
}

// NewKVPruner returns a KVPruner using the given client and mount.
func NewKVPruner(client *Client, mount string) *KVPruner {
	if mount == "" {
		mount = "secret"
	}
	return &KVPruner{client: client, mount: mount}
}

// Prune retains only the `keep` most recent versions of the secret at path,
// permanently destroying older versions.
func (p *KVPruner) Prune(ctx context.Context, path string, keep int) (*PruneResult, error) {
	if keep < 1 {
		return nil, fmt.Errorf("keep must be at least 1")
	}

	metaURL := fmt.Sprintf("%s/v1/%s/metadata/%s", p.client.Address, p.mount, path)
	resp, err := p.client.RawRequestWithContext(ctx, http.MethodGet, metaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching metadata", resp.StatusCode)
	}

	var meta struct {
		Data struct {
			Versions map[string]struct {
				Destroyed bool `json:"destroyed"`
			} `json:"versions"`
		} `json:"data"`
	}
	if err := decodeJSON(resp.Body, &meta); err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}

	versions := sortedVersionInts(meta.Data.Versions)
	result := &PruneResult{Key: path}

	if len(versions) <= keep {
		result.Skipped = len(versions)
		return result, nil
	}

	toDestroy := versions[:len(versions)-keep]
	result.Skipped = keep

	destroyURL := fmt.Sprintf("%s/v1/%s/destroy/%s", p.client.Address, p.mount, path)
	body := map[string]interface{}{"versions": toDestroy}
	dresp, err := p.client.RawRequestWithContext(ctx, http.MethodPut, destroyURL, body)
	if err != nil {
		return nil, fmt.Errorf("destroy versions: %w", err)
	}
	defer dresp.Body.Close()

	if dresp.StatusCode != http.StatusNoContent && dresp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d destroying versions", dresp.StatusCode)
	}

	result.Pruned = len(toDestroy)
	return result, nil
}
