package vault

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// BulkPutResult holds the outcome of a single write operation within a bulk put.
type BulkPutResult struct {
	Path    string
	Success bool
	Error   error
}

// KVBulkWriter writes multiple KV secrets in a single operation.
type KVBulkWriter struct {
	client *vaultapi.Client
	mount  string
}

// NewKVBulkWriter creates a new KVBulkWriter with the given Vault client and mount path.
// If mount is empty, it defaults to "secret".
func NewKVBulkWriter(client *vaultapi.Client, mount string) *KVBulkWriter {
	if mount == "" {
		mount = "secret"
	}
	return &KVBulkWriter{client: client, mount: mount}
}

// BulkPut writes each path→data entry to Vault KV v2.
// It attempts all writes and returns a slice of BulkPutResult, one per entry.
// An error is returned only if the entries map is nil or empty.
func (w *KVBulkWriter) BulkPut(ctx context.Context, entries map[string]map[string]string) ([]BulkPutResult, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries provided for bulk put")
	}

	results := make([]BulkPutResult, 0, len(entries))

	for path, data := range entries {
		path = strings.TrimPrefix(path, "/")
		apiPath := fmt.Sprintf("%s/data/%s", w.mount, path)

		body := map[string]interface{}{
			"data": toInterfaceMap(data),
		}

		resp, err := w.client.Logical().WriteWithContext(ctx, apiPath, body)
		if err != nil {
			results = append(results, BulkPutResult{Path: path, Success: false, Error: err})
			continue
		}

		if resp == nil {
			results = append(results, BulkPutResult{
				Path:    path,
				Success: false,
				Error:   fmt.Errorf("empty response from Vault for path %q", path),
			})
			continue
		}

		results = append(results, BulkPutResult{Path: path, Success: true})
	}

	return results, nil
}

// BulkPutStatus summarises the results of a BulkPut call.
type BulkPutStatus struct {
	Succeeded []string
	Failed    map[string]error
}

// Summarise converts a slice of BulkPutResult into a BulkPutStatus for easy reporting.
func Summarise(results []BulkPutResult) BulkPutStatus {
	status := BulkPutStatus{
		Failed: make(map[string]error),
	}
	for _, r := range results {
		if r.Success {
			status.Succeeded = append(status.Succeeded, r.Path)
		} else {
			status.Failed[r.Path] = r.Error
		}
	}
	return status
}

// toInterfaceMap converts map[string]string to map[string]interface{} as required by the Vault API.
func toInterfaceMap(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// httpStatusFromError attempts to extract an HTTP status code from a Vault API error.
// Returns 0 if the error is not a *vaultapi.ResponseError.
func httpStatusFromError(err error) int {
	if re, ok := err.(*vaultapi.ResponseError); ok {
		return re.StatusCode
	}
	return http.StatusInternalServerError
}
