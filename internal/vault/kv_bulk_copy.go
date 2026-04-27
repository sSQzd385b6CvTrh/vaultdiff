package vault

import (
	"context"
	"fmt"
	"net/http"
)

// BulkCopyResult holds the outcome for a single key copy operation.
type BulkCopyResult struct {
	Source      string
	Destination string
	Err         error
}

// KVBulkCopier copies multiple KV secrets from source paths to destination paths.
type KVBulkCopier struct {
	client *Client
	mount  string
}

// NewKVBulkCopier returns a new KVBulkCopier.
func NewKVBulkCopier(client *Client, mount string) *KVBulkCopier {
	if mount == "" {
		mount = "secret"
	}
	return &KVBulkCopier{client: client, mount: mount}
}

// CopyPair represents a source→destination copy request.
type CopyPair struct {
	Source      string
	Destination string
}

// Copy performs bulk copy of secrets. Each source is read and written to its
// corresponding destination. Results are returned for all pairs regardless of
// individual failures.
func (b *KVBulkCopier) Copy(ctx context.Context, pairs []CopyPair) []BulkCopyResult {
	results := make([]BulkCopyResult, 0, len(pairs))
	for _, p := range pairs {
		err := b.copySingle(ctx, p.Source, p.Destination)
		results = append(results, BulkCopyResult{
			Source:      p.Source,
			Destination: p.Destination,
			Err:         err,
		})
	}
	return results
}

func (b *KVBulkCopier) copySingle(ctx context.Context, src, dst string) error {
	getURL := fmt.Sprintf("%s/v1/%s/data/%s", b.client.Address, b.mount, src)
	resp, err := b.client.RawRequestWithContext(ctx, http.MethodGet, getURL, nil)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("source not found: %s", src)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d reading %s", resp.StatusCode, src)
	}

	var envelope map[string]interface{}
	if err := decodeJSON(resp.Body, &envelope); err != nil {
		return fmt.Errorf("decode %s: %w", src, err)
	}
	data, _ := envelope["data"].(map[string]interface{})
	kvData, _ := data["data"].(map[string]interface{})

	putURL := fmt.Sprintf("%s/v1/%s/data/%s", b.client.Address, b.mount, dst)
	body := map[string]interface{}{"data": kvData}
	putResp, err := b.client.RawRequestWithContext(ctx, http.MethodPost, putURL, body)
	if err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d writing %s", putResp.StatusCode, dst)
	}
	return nil
}
