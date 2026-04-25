package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/thrasher-corp/vaultdiff/internal/vault/client"
)

// SchemaField describes an expected key in a KV secret.
type SchemaField struct {
	Key      string
	Required bool
}

// SchemaResult holds the outcome of a schema validation.
type SchemaResult struct {
	Path    string
	Missing []string
	Extra   []string
	Valid   bool
}

// KVSchemaValidator validates a KV secret against an expected set of keys.
type KVSchemaValidator struct {
	client *client.Client
	mount  string
}

// NewKVSchemaValidator creates a new KVSchemaValidator.
func NewKVSchemaValidator(c *client.Client, mount string) *KVSchemaValidator {
	if mount == "" {
		mount = "secret"
	}
	return &KVSchemaValidator{client: c, mount: mount}
}

// Validate fetches the secret at path and compares its keys against fields.
func (v *KVSchemaValidator) Validate(path string, fields []SchemaField) (*SchemaResult, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", v.client.Address, v.mount, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", v.client.Token)

	resp, err := v.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	actual := body.Data.Data
	required := map[string]bool{}
	for _, f := range fields {
		if f.Required {
			required[f.Key] = true
		}
	}

	var missing, extra []string
	for k := range required {
		if _, ok := actual[k]; !ok {
			missing = append(missing, k)
		}
	}
	allowed := map[string]bool{}
	for _, f := range fields {
		allowed[f.Key] = true
	}
	for k := range actual {
		if !allowed[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)

	return &SchemaResult{
		Path:    path,
		Missing: missing,
		Extra:   extra,
		Valid:   len(missing) == 0,
	}, nil
}
