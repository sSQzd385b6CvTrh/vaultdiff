package vault

import (
	"context"
	"fmt"
	"net/http"
)

// KVPromoter copies a secret from one path to another (e.g. staging -> prod)
// and optionally deletes the source after a successful copy.
type KVPromoter struct {
	client    *Client
	mount     string
	deletesrc bool
}

// NewKVPromoter returns a KVPromoter for the given mount.
// If mount is empty it defaults to "secret".
func NewKVPromoter(c *Client, mount string, deleteSrc bool) *KVPromoter {
	if mount == "" {
		mount = "secret"
	}
	return &KVPromoter{client: c, mount: mount, deletesSrc: deleteSrc}
}

// PromoteResult holds the data written to the destination path.
type PromoteResult struct {
	SourcePath string
	DestPath   string
	Data       map[string]interface{}
}

// Promote reads the latest version of srcPath and writes it to dstPath.
func (p *KVPromoter) Promote(ctx context.Context, srcPath, dstPath string) (*PromoteResult, error) {
	// Read source
	readURL := fmt.Sprintf("%s/v1/%s/data/%s", p.client.Address, p.mount, srcPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, readURL, nil)
	if err != nil {
		return nil, fmt.Errorf("promote: build read request: %w", err)
	}
	p.client.applyHeaders(req)

	resp, err := p.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("promote: read source: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("promote: source path %q not found", srcPath)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("promote: unexpected status reading source: %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := decodeJSON(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("promote: decode source: %w", err)
	}

	// Write destination
	writeURL := fmt.Sprintf("%s/v1/%s/data/%s", p.client.Address, p.mount, dstPath)
	payload := map[string]interface{}{"data": body.Data.Data}
	wreq, err := newJSONRequest(ctx, http.MethodPost, writeURL, payload)
	if err != nil {
		return nil, fmt.Errorf("promote: build write request: %w", err)
	}
	p.client.applyHeaders(wreq)

	wresp, err := p.client.HTTP.Do(wreq)
	if err != nil {
		return nil, fmt.Errorf("promote: write destination: %w", err)
	}
	defer wresp.Body.Close()

	if wresp.StatusCode != http.StatusOK && wresp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("promote: unexpected status writing destination: %d", wresp.StatusCode)
	}

	if p.deletesSrc {
		delURL := fmt.Sprintf("%s/v1/%s/data/%s", p.client.Address, p.mount, srcPath)
		dreq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, delURL, nil)
		p.client.applyHeaders(dreq)
		p.client.HTTP.Do(dreq) // best-effort
	}

	return &PromoteResult{
		SourcePath: srcPath,
		DestPath:   dstPath,
		Data:       body.Data.Data,
	}, nil
}
