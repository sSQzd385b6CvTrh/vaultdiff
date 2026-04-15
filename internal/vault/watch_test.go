package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

func watchKVResponse(version int, data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{
				"version": float64(version),
			},
		},
	}
}

func newWatchServer(t *testing.T, responses []map[string]interface{}) (*httptest.Server, *int32) {
	t.Helper()
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := int(atomic.AddInt32(&callCount, 1)) - 1
		if idx >= len(responses) {
			idx = len(responses) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses[idx])
	}))
	return server, &callCount
}

func TestWatch_DetectsVersionChange(t *testing.T) {
	responses := []map[string]interface{}{
		watchKVResponse(1, map[string]interface{}{"key": "v1"}),
		watchKVResponse(2, map[string]interface{}{"key": "v2"}),
	}
	server, _ := newWatchServer(t, responses)
	defer server.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	w := NewWatcher(client, "secret", 20*time.Millisecond)
	ch := make(chan WatchEvent, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	go w.Watch(ctx, "myapp/config", ch)

	var events []WatchEvent
	for e := range ch {
		if e.Err != nil {
			continue
		}
		events = append(events, e)
		if len(events) >= 2 {
			cancel()
			break
		}
	}

	if len(events) < 2 {
		t.Fatalf("expected 2 change events, got %d", len(events))
	}
	if events[0].Version != 1 {
		t.Errorf("first event version = %d, want 1", events[0].Version)
	}
	if events[1].Version != 2 {
		t.Errorf("second event version = %d, want 2", events[1].Version)
	}
}

func TestNewWatcher_Defaults(t *testing.T) {
	client, _ := vaultapi.NewClient(vaultapi.DefaultConfig())
	w := NewWatcher(client, "", 0)
	if w.mount != "secret" {
		t.Errorf("mount = %q, want \"secret\"", w.mount)
	}
	if w.interval != 30*time.Second {
		t.Errorf("interval = %v, want 30s", w.interval)
	}
}
