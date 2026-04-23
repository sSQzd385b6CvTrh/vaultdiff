package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func checksumKVResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{"version": 1},
		},
	}
}

func newChecksumServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestKVChecksum_Success(t *testing.T) {
	payload := checksumKVResponse(map[string]interface{}{"foo": "bar", "baz": "qux"})
	srv := newChecksumServer(t, http.StatusOK, payload)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := vaultapi.NewClient(cfg)

	checksummer := NewKVChecksummer(c, "secret")
	result, err := checksummer.Checksum("myapp/config", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Path != "myapp/config" {
		t.Errorf("expected path myapp/config, got %s", result.Path)
	}
	if result.Version != 1 {
		t.Errorf("expected version 1, got %d", result.Version)
	}
	if len(result.Sum) != 64 {
		t.Errorf("expected 64-char hex digest, got %d chars", len(result.Sum))
	}
}

func TestKVChecksum_Deterministic(t *testing.T) {
	data := map[string]interface{}{"z": "last", "a": "first", "m": "middle"}
	sum1, err := checksumData(data)
	if err != nil {
		t.Fatal(err)
	}
	sum2, err := checksumData(data)
	if err != nil {
		t.Fatal(err)
	}
	if sum1 != sum2 {
		t.Errorf("checksums not deterministic: %s vs %s", sum1, sum2)
	}
}

func TestKVChecksum_NotFound(t *testing.T) {
	srv := newChecksumServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, _ := vaultapi.NewClient(cfg)

	checksummer := NewKVChecksummer(c, "secret")
	_, err := checksummer.Checksum("missing/path", 1)
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewKVChecksummer_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	c, _ := vaultapi.NewClient(cfg)
	ch := NewKVChecksummer(c, "")
	if ch.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", ch.mount)
	}
}
