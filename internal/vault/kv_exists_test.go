package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func kvExistsResponse(version int, destroyed bool) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"metadata": map[string]interface{}{
				"version":   version,
				"destroyed": destroyed,
			},
		},
	}
}

func newExistsServer(statusCode int, body interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestKVExists_Found(t *testing.T) {
	ts := newExistsServer(http.StatusOK, kvExistsResponse(3, false))
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	checker := NewKVExistenceChecker(client, "secret")
	result, err := checker.Check(context.Background(), "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Exists {
		t.Error("expected Exists=true")
	}
	if result.Version != 3 {
		t.Errorf("expected version 3, got %d", result.Version)
	}
	if result.Destroyed {
		t.Error("expected Destroyed=false")
	}
}

func TestKVExists_NotFound(t *testing.T) {
	ts := newExistsServer(http.StatusNotFound, nil)
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	checker := NewKVExistenceChecker(client, "secret")
	result, err := checker.Check(context.Background(), "myapp/missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Exists {
		t.Error("expected Exists=false")
	}
}

func TestKVExists_UnexpectedStatus(t *testing.T) {
	ts := newExistsServer(http.StatusInternalServerError, nil)
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := vaultapi.NewClient(cfg)

	checker := NewKVExistenceChecker(client, "secret")
	_, err := checker.Check(context.Background(), "myapp/config")
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}

func TestNewKVExistenceChecker_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	client, _ := vaultapi.NewClient(cfg)
	checker := NewKVExistenceChecker(client, "")
	if checker.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", checker.mount)
	}
}
