package vault

import (
	"fmt"
	"net/http"
)

// BulkMoveResult holds the outcome of a single key move.
type BulkMoveResult struct {
	Source string
	Dest   string
	Err    error
}

// KVBulkMover moves multiple KV secrets atomically (copy then delete).
type KVBulkMover struct {
	client *Client
	mount  string
}

// NewKVBulkMover creates a KVBulkMover with the given client and mount.
func NewKVBulkMover(client *Client, mount string) *KVBulkMover {
	if mount == "" {
		mount = "secret"
	}
	return &KVBulkMover{client: client, mount: mount}
}

// MovePair represents a source→destination pair.
type MovePair struct {
	Source string
	Dest   string
}

// Move copies each source to its destination then deletes the source.
// It returns one BulkMoveResult per pair.
func (m *KVBulkMover) Move(pairs []MovePair) []BulkMoveResult {
	results := make([]BulkMoveResult, 0, len(pairs))
	for _, p := range pairs {
		res := BulkMoveResult{Source: p.Source, Dest: p.Dest}
		if p.Source == p.Dest {
			res.Err = fmt.Errorf("source and destination are identical: %s", p.Source)
			results = append(results, res)
			continue
		}
		// Read source
		readURL := fmt.Sprintf("%s/v1/%s/data/%s", m.client.Address, m.mount, p.Source)
		readResp, err := m.client.HTTPClient.Get(readURL)
		if err != nil {
			res.Err = fmt.Errorf("read %s: %w", p.Source, err)
			results = append(results, res)
			continue
		}
		defer readResp.Body.Close()
		if readResp.StatusCode == http.StatusNotFound {
			res.Err = fmt.Errorf("source not found: %s", p.Source)
			results = append(results, res)
			continue
		}
		if readResp.StatusCode != http.StatusOK {
			res.Err = fmt.Errorf("unexpected status reading %s: %d", p.Source, readResp.StatusCode)
			results = append(results, res)
			continue
		}
		// Copy via KVCopier
		copier := NewKVCopier(m.client, m.mount)
		if err := copier.Copy(p.Source, p.Dest); err != nil {
			res.Err = fmt.Errorf("copy %s→%s: %w", p.Source, p.Dest, err)
			results = append(results, res)
			continue
		}
		// Delete source
		deleter := NewKVDeleter(m.client, m.mount)
		if err := deleter.Delete(p.Source); err != nil {
			res.Err = fmt.Errorf("delete %s: %w", p.Source, err)
		}
		results = append(results, res)
	}
	return results
}
