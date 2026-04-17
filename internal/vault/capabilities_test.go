package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func capabilitiesResponse(data map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

func TestCheckSelf_Success(t *testing.T) {
	paths := []string{"secret/data/foo", "secret/data/bar"}
	response := map[string]interface{}{
		"secret/data/foo": []interface{}{"read", "list"},
		"secret/data/bar": []interface{}{"deny"},
	}

	ts := httptest.NewServer(capabilitiesResponse(response))
	defer ts.Close()

	cfg := api.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := api.NewClient(cfg)

	checker := NewCapabilityChecker(client)
	results, err := checker.CheckSelf(context.Background(), paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Path != "secret/data/foo" {
		t.Errorf("expected path secret/data/foo, got %s", results[0].Path)
	}
	if len(results[0].Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(results[0].Capabilities))
	}
}

func TestCheckSelf_Forbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	cfg := api.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := api.NewClient(cfg)

	checker := NewCapabilityChecker(client)
	_, err := checker.CheckSelf(context.Background(), []string{"secret/data/foo"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckSelf_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := api.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := api.NewClient(cfg)

	checker := NewCapabilityChecker(client)
	_, err := checker.CheckSelf(context.Background(), []string{"secret/data/foo"})
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}
