package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func kvLatestResponse(version int, data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{
				"version": version,
			},
		},
	}
}

func newKVLatestServer(t *testing.T, path string, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestGetLatest_Success(t *testing.T) {
	payload := kvLatestResponse(7, map[string]interface{}{"api_key": "abc123", "region": "us-east-1"})
	srv := newKVLatestServer(t, "/v1/secret/data/myapp/config", http.StatusOK, payload)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	reader := NewKVLatestReader(client, "secret")
	lv, err := reader.GetLatest(context.Background(), "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lv.Version != 7 {
		t.Errorf("expected version 7, got %d", lv.Version)
	}
	if lv.Data["api_key"] != "abc123" {
		t.Errorf("expected api_key=abc123, got %s", lv.Data["api_key"])
	}
}

func TestGetLatest_NotFound(t *testing.T) {
	srv := newKVLatestServer(t, "", http.StatusNotFound, nil)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	reader := NewKVLatestReader(client, "secret")
	_, err := reader.GetLatest(context.Background(), "missing/path")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestGetLatest_UnexpectedStatus(t *testing.T) {
	srv := newKVLatestServer(t, "", http.StatusInternalServerError, nil)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	reader := NewKVLatestReader(client, "secret")
	_, err := reader.GetLatest(context.Background(), "some/path")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestNewKVLatestReader_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	reader := NewKVLatestReader(client, "")
	if reader.mount != "secret" {
		t.Errorf("expected default mount 'secret', got '%s'", reader.mount)
	}
}
