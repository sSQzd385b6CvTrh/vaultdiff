package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

func kvStatMetadataResponse(currentVersion, totalVersions int, destroyed bool) map[string]interface{} {
	versions := map[string]interface{}{}
	for i := 1; i <= totalVersions; i++ {
		key := fmt.Sprintf("%d", i)
		versions[key] = map[string]interface{}{
			"destroyed":    i == currentVersion && destroyed,
			"deletion_time": "",
		}
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"current_version": currentVersion,
			"oldest_version":  1,
			"created_time":    time.Now().UTC().Format(time.RFC3339),
			"updated_time":    time.Now().UTC().Format(time.RFC3339),
			"versions":        versions,
		},
	}
}

func newKVStatServer(t *testing.T, path string, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestKVStat_Success(t *testing.T) {
	payload := kvStatMetadataResponse(3, 3, false)
	ts := newKVStatServer(t, "myapp/config", http.StatusOK, payload)
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	c, _ := vaultapi.NewClient(cfg)

	statter := NewKVStatter(c, "secret")
	stat, err := statter.Stat(context.Background(), "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stat.CurrentVersion != 3 {
		t.Errorf("expected version 3, got %d", stat.CurrentVersion)
	}
	if stat.TotalVersions != 3 {
		t.Errorf("expected 3 total versions, got %d", stat.TotalVersions)
	}
	if stat.Destroyed {
		t.Error("expected not destroyed")
	}
}

func TestKVStat_NotFound(t *testing.T) {
	ts := newKVStatServer(t, "missing", http.StatusNotFound, nil)
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	c, _ := vaultapi.NewClient(cfg)

	statter := NewKVStatter(c, "secret")
	_, err := statter.Stat(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestNewKVStatter_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	c, _ := vaultapi.NewClient(cfg)
	s := NewKVStatter(c, "")
	if s.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", s.mount)
	}
}
