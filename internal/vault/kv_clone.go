package vault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/api"
)

// KVCloner copies all key-value pairs from one path to another, optionally
// across different mounts.
type KVCloner struct {
	client *api.Client
	mount  string
}

// NewKVCloner returns a KVCloner using the provided Vault client.
// mount defaults to "secret" if empty.
func NewKVCloner(client *api.Client, mount string) *KVCloner {
	if mount == "" {
		mount = "secret"
	}
	return &KVCloner{client: client, mount: mount}
}

// Clone reads the latest version of srcPath and writes all its data to
// dstPath under dstMount. If dstMount is empty the same mount is used.
func (c *KVCloner) Clone(srcPath, dstPath, dstMount string) (map[string]interface{}, error) {
	if dstMount == "" {
		dstMount = c.mount
	}

	readPath := fmt.Sprintf("%s/data/%s", c.mount, srcPath)
	secret, err := c.client.Logical().Read(readPath)
	if err != nil {
		return nil, fmt.Errorf("clone read %s: %w", srcPath, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("clone: source path %q not found (status %d)", srcPath, http.StatusNotFound)
	}

	dataRaw, ok := secret.Data["data"]
	if !ok {
		return nil, fmt.Errorf("clone: missing data key in source response")
	}
	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("clone: unexpected data type in source response")
	}

	writePath := fmt.Sprintf("%s/data/%s", dstMount, dstPath)
	payload := map[string]interface{}{"data": data}
	result, err := c.client.Logical().Write(writePath, payload)
	if err != nil {
		return nil, fmt.Errorf("clone write %s: %w", dstPath, err)
	}
	if result == nil {
		return nil, fmt.Errorf("clone: empty response writing to %q", dstPath)
	}
	return result.Data, nil
}
