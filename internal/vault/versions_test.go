package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func metadataResponse(versions map[string]interface{}) []byte {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"versions": versions,
		},
	}
	b, _ := json.Marshal(body)
	return b
}

func TestListVersions_Success(t *testing.T) {
	versions := map[string]interface{}{
		"1": map[string]interface{}{
			"created_time":  "2024-01-01T00:00:00Z",
			"deletion_time": "",
			"destroyed":     false,
		},
		"2": map[string]interface{}{
			"created_time":  "2024-02-01T00:00:00Z",
			"deletion_time": "",
			"destroyed":     false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(metadataResponse(versions))
	}))
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	f := &Fetcher{client: client}
	vl := NewVersionLister(f)

	metas, err := vl.ListVersions(context.Background(), "secret", "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(metas) != 2 {
		t.Errorf("expected 2 versions, got %d", len(metas))
	}
}

func TestListVersions_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	f := &Fetcher{client: client}
	vl := NewVersionLister(f)

	_, err := vl.ListVersions(context.Background(), "secret", "missing/path")
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
}

func TestListVersions_MissingVersionsKey(t *testing.T) {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"other_key": "value",
		},
	}
	b, _ := json.Marshal(body)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}))
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	f := &Fetcher{client: client}
	vl := NewVersionLister(f)

	_, err := vl.ListVersions(context.Background(), "secret", "myapp/config")
	if err == nil {
		t.Fatal("expected error for missing versions key, got nil")
	}
}
