package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func kvValidateResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"data": data}
}

func newValidateServer(t *testing.T, status int, body map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestKVValidate_Success(t *testing.T) {
	body := kvValidateResponse(map[string]interface{}{
		"foo": "bar",
		"metadata": map[string]interface{}{
			"destroyed":     false,
			"deletion_time": "",
		},
	})
	server := newValidateServer(t, http.StatusOK, body)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	v := NewKVValidator(client, "secret")
	res, err := v.Validate("myapp/config", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Valid {
		t.Errorf("expected valid, got reason: %s", res.Reason)
	}
}

func TestKVValidate_Destroyed(t *testing.T) {
	body := kvValidateResponse(map[string]interface{}{
		"metadata": map[string]interface{}{
			"destroyed":     true,
			"deletion_time": "",
		},
	})
	server := newValidateServer(t, http.StatusOK, body)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	v := NewKVValidator(client, "secret")
	res, err := v.Validate("myapp/config", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Error("expected invalid for destroyed version")
	}
	if res.Reason != "version is destroyed" {
		t.Errorf("unexpected reason: %s", res.Reason)
	}
}

func TestKVValidate_NotFound(t *testing.T) {
	server := newValidateServer(t, http.StatusNotFound, nil)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)

	v := NewKVValidator(client, "secret")
	res, err := v.Validate("missing/path", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Error("expected invalid for missing secret")
	}
	if res.Reason != "secret not found" {
		t.Errorf("unexpected reason: %s", res.Reason)
	}
}

func TestNewKVValidator_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	v := NewKVValidator(client, "")
	if v.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", v.mount)
	}
}
