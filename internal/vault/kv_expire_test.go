package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func kvExpireMetadataResponse(createdTime, deletionTime string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"current_version": 2,
			"versions": map[string]interface{}{
				"2": map[string]interface{}{
					"created_time":  createdTime,
					"deletion_time": deletionTime,
					"destroyed":     false,
				},
			},
		},
	}
}

func newExpireServer(t *testing.T, status int, body map[string]interface{}) (*httptest.Server, *KVExpirer) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(ts.Close)
	client := newTestClient(t, ts.URL)
	return ts, NewKVExpirer(client, "secret")
}

func TestCheckExpiry_WithDeletionTime(t *testing.T) {
	future := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339Nano)
	_, expirer := newExpireServer(t, http.StatusOK, kvExpireMetadataResponse("2024-01-01T00:00:00Z", future))

	info, err := expirer.CheckExpiry("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != 2 {
		t.Errorf("expected version 2, got %d", info.Version)
	}
	if info.Expired {
		t.Error("expected secret to not be expired")
	}
	if info.TTL <= 0 {
		t.Error("expected positive TTL")
	}
}

func TestCheckExpiry_NoDeletionTime(t *testing.T) {
	_, expirer := newExpireServer(t, http.StatusOK, kvExpireMetadataResponse("2024-01-01T00:00:00Z", ""))

	info, err := expirer.CheckExpiry("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.ExpiresAt.IsZero() {
		t.Error("expected zero ExpiresAt when no deletion_time set")
	}
}

func TestCheckExpiry_NotFound(t *testing.T) {
	_, expirer := newExpireServer(t, http.StatusNotFound, map[string]interface{}{})

	_, err := expirer.CheckExpiry("missing/path")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestNewKVExpirer_DefaultMount(t *testing.T) {
	client, _ := NewClient("", "", "")
	e := NewKVExpirer(client, "")
	if e.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", e.mount)
	}
}
