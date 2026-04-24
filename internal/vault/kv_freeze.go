package vault

import (
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
)

// KVFreezeResult holds the result of a freeze or unfreeze operation.
type KVFreezeResult struct {
	Path    string
	Frozen  bool
	Message string
}

// KVFreezer marks a KV secret as frozen by setting a custom metadata flag.
type KVFreezer struct {
	client *vaultapi.Client
	mount  string
}

// NewKVFreezer creates a new KVFreezer. Defaults mount to "secret".
func NewKVFreezer(client *vaultapi.Client, mount string) *KVFreezer {
	if mount == "" {
		mount = "secret"
	}
	return &KVFreezer{client: client, mount: mount}
}

// Freeze sets the frozen=true custom metadata on the given secret path.
func (f *KVFreezer) Freeze(path string) (*KVFreezeResult, error) {
	return f.setFrozen(path, true)
}

// Unfreeze clears the frozen flag from the given secret path.
func (f *KVFreezer) Unfreeze(path string) (*KVFreezeResult, error) {
	return f.setFrozen(path, false)
}

func (f *KVFreezer) setFrozen(path string, frozen bool) (*KVFreezeResult, error) {
	metaPath := fmt.Sprintf("/v1/%s/metadata/%s", f.mount, path)
	body := map[string]interface{}{
		"custom_metadata": map[string]string{
			"frozen": fmt.Sprintf("%v", frozen),
		},
	}
	resp, err := f.client.RawRequest(f.client.NewRequest(http.MethodPost, metaPath).WithJSONBody(body))
	if err != nil {
		return nil, fmt.Errorf("freeze request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for path %s", resp.StatusCode, path)
	}
	msg := "unfrozen"
	if frozen {
		msg = "frozen"
	}
	return &KVFreezeResult{Path: path, Frozen: frozen, Message: fmt.Sprintf("%s successfully %s", path, msg)}, nil
}

// IsFrozen checks whether a secret path has the frozen custom metadata flag set to true.
func (f *KVFreezer) IsFrozen(path string) (bool, error) {
	metaPath := fmt.Sprintf("%s/metadata/%s", f.mount, path)
	secret, err := f.client.Logical().Read(metaPath)
	if err != nil {
		return false, fmt.Errorf("failed to read metadata: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return false, fmt.Errorf("secret not found: %s", path)
	}
	cm, ok := secret.Data["custom_metadata"]
	if !ok {
		return false, nil
	}
	meta, ok := cm.(map[string]interface{})
	if !ok {
		return false, nil
	}
	return meta["frozen"] == "true", nil
}
