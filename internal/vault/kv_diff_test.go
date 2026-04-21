package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func kvDiffResponse(data map[string]interface{}) []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
		},
	})
	return body
}

func newKVDiffServer(srcData, dstData map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/secret/data/src/app" {
			w.WriteHeader(http.StatusOK)
			w.Write(kvDiffResponse(srcData))
		} else if r.URL.Path == "/v1/secret/data/dst/app" {
			w.WriteHeader(http.StatusOK)
			w.Write(kvDiffResponse(dstData))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{}`))
		}
	}))
}

func TestKVDiff_Success(t *testing.T) {
	srcData := map[string]interface{}{"key": "value1", "shared": "same"}
	dstData := map[string]interface{}{"key": "value2", "shared": "same"}
	server := newKVDiffServer(srcData, dstData)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	differ := NewKVDiffer(client, "secret")
	result, err := differ.Diff("src/app", "dst/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SourceData["key"] != "value1" {
		t.Errorf("expected source key=value1, got %v", result.SourceData["key"])
	}
	if result.TargetData["key"] != "value2" {
		t.Errorf("expected target key=value2, got %v", result.TargetData["key"])
	}
}

func TestKVDiff_SourceNotFound(t *testing.T) {
	server := newKVDiffServer(nil, nil)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	differ := NewKVDiffer(client, "secret")
	_, err := differ.Diff("missing/path", "dst/app")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestNewKVDiffer_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	d := NewKVDiffer(client, "")
	if d.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", d.mount)
	}
}
