package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func touchKVGetResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": map[string]interface{}{
				"username": "admin",
				"password": "s3cr3t",
			},
		},
	}
}

func touchKVPutResponse(version int) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"version": version,
		},
	}
}

func newTouchServer(t *testing.T, getStatus, putStatus int, newVersion int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.WriteHeader(getStatus)
			if getStatus == http.StatusOK {
				json.NewEncoder(w).Encode(touchKVGetResponse())
			}
			return
		}
		w.WriteHeader(putStatus)
		if putStatus == http.StatusOK {
			json.NewEncoder(w).Encode(touchKVPutResponse(newVersion))
		}
	}))
}

func TestKVTouch_Success(t *testing.T) {
	srv := newTouchServer(t, http.StatusOK, http.StatusOK, 4)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	toucher := NewKVToucher(client, "secret")
	version, err := toucher.Touch(context.Background(), "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != 4 {
		t.Errorf("expected version 4, got %d", version)
	}
}

func TestKVTouch_NotFound(t *testing.T) {
	srv := newTouchServer(t, http.StatusNotFound, http.StatusOK, 0)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	toucher := NewKVToucher(client, "secret")
	_, err := toucher.Touch(context.Background(), "missing/path")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
}

func TestKVTouch_WriteError(t *testing.T) {
	srv := newTouchServer(t, http.StatusOK, http.StatusInternalServerError, 0)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	toucher := NewKVToucher(client, "secret")
	_, err := toucher.Touch(context.Background(), "myapp/config")
	if err == nil {
		t.Fatal("expected error on write failure, got nil")
	}
}

func TestNewKVToucher_DefaultMount(t *testing.T) {
	client, _ := api.NewClient(api.DefaultConfig())
	toucher := NewKVToucher(client, "")
	if toucher.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", toucher.mount)
	}
}
