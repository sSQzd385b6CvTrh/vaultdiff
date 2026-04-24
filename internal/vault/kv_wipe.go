package vault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// KVWiper deletes all versions and metadata for a list of KV secrets.
type KVWiper struct {
	client *api.Client
	mount  string
}

// NewKVWiper returns a new KVWiper using the provided Vault client.
// If mount is empty it defaults to "secret".
func NewKVWiper(client *api.Client, mount string) *KVWiper {
	if mount == "" {
		mount = "secret"
	}
	return &KVWiper{client: client, mount: mount}
}

// WipeResult holds the outcome for a single path.
type WipeResult struct {
	Path    string
	Wiped   bool
	Message string
}

// Wipe permanently deletes all versions and metadata for each path.
// It calls DELETE /v1/<mount>/metadata/<path> which removes everything.
func (w *KVWiper) Wipe(paths []string) ([]WipeResult, error) {
	results := make([]WipeResult, 0, len(paths))
	for _, p := range paths {
		endpoint := fmt.Sprintf("%s/metadata/%s", w.mount, p)
		resp, err := w.client.Logical().Delete(endpoint)
		if err != nil {
			results = append(results, WipeResult{Path: p, Wiped: false, Message: err.Error()})
			continue
		}
		code := http.StatusNoContent
		if resp != nil {
			_ = resp
		}
		_ = code
		results = append(results, WipeResult{Path: p, Wiped: true, Message: "wiped"})
	}
	return results, nil
}
