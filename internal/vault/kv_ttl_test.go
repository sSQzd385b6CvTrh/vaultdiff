package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

func kvTTLMetadataResponse(deleteVersionAfter string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"delete_version_after": deleteVersionAfter,
			"max_versions":         float64(10),
		},
	}
}

func newTTLServer(t *testing.T, statusCode int, body map[string]interface{}) (*httptest.Server, *api.Client) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(body)
	}))
	cfg := api.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := api.NewClient(cfg)
	t.Cleanup(ts.Close)
	return ts, client
}

func TestGetTTL_WithTTL(t *testing.T) {
	_, client := newTTLServer(t, 200, kvTTLMetadataResponse("72h"))
	reader := NewKVTTLReader(client, "secret")
	info, err := reader.GetTTL("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasTTL {
		t.Error("expected HasTTL to be true")
	}
	if info.Remaining.Hours() < 71 {
		t.Errorf("expected remaining ~72h, got %v", info.Remaining)
	}
}

func TestGetTTL_NoTTL(t *testing.T) {
	_, client := newTTLServer(t, 200, kvTTLMetadataResponse("0s"))
	reader := NewKVTTLReader(client, "secret")
	info, err := reader.GetTTL("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.HasTTL {
		t.Error("expected HasTTL to be false")
	}
}

func TestGetTTL_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()
	cfg := api.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := api.NewClient(cfg)
	reader := NewKVTTLReader(client, "secret")
	_, err := reader.GetTTL("missing/path")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewKVTTLReader_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	reader := NewKVTTLReader(client, "")
	if reader.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", reader.mount)
	}
}

func TestGetTTL_InvalidDuration(t *testing.T) {
	_, client := newTTLServer(t, 200, kvTTLMetadataResponse("notaduration"))
	reader := NewKVTTLReader(client, "secret")
	_, err := reader.GetTTL("myapp/config")
	if err == nil || !strings.Contains(err.Error(), "parsing") {
		t.Errorf("expected parse error, got %v", err)
	}
}
