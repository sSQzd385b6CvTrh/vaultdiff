package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func kvWatchKeysResponse(version int) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": map[string]interface{}{"key": "value"},
			"metadata": map[string]interface{}{"version": float64(version)},
		},
	}
}

func newKVWatchKeysServer(version int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(kvWatchKeysResponse(version))
	}))
}

func TestKVKeyWatcher_DetectsChange(t *testing.T) {
	version := 1
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(kvWatchKeysResponse(version))
	}))
	defer ts.Close()

	client := newTestVaultClient(t, ts.URL)
	watcher := NewKVKeyWatcher(client, "secret", 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	ch, err := watcher.WatchKeys(ctx, []string{"myapp/config"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First tick: version=1 (change from 0)
	select {
	case result := <-ch:
		if result.Version != 1 {
			t.Errorf("expected version 1, got %d", result.Version)
		}
		if !result.Changed {
			t.Error("expected Changed=true")
		}
		if result.Path != "myapp/config" {
			t.Errorf("expected path myapp/config, got %s", result.Path)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for watch result")
	}
}

func TestKVKeyWatcher_NoChangeWhenVersionSame(t *testing.T) {
	ts := newKVWatchKeysServer(3)
	defer ts.Close()

	client := newTestVaultClient(t, ts.URL)
	watcher := NewKVKeyWatcher(client, "secret", 40*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ch, err := watcher.WatchKeys(ctx, []string{"myapp/config"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Drain the first change (version 3 vs known 0)
	<-ch

	// No further changes expected since version stays at 3
	select {
	case r, ok := <-ch:
		if ok {
			t.Errorf("unexpected change received: %+v", r)
		}
	case <-ctx.Done():
		// expected: no further results
	}
}

func TestNewKVKeyWatcher_Defaults(t *testing.T) {
	client := newTestVaultClient(t, "http://localhost:8200")
	w := NewKVKeyWatcher(client, "", 0)
	if w.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", w.mount)
	}
	if w.interval != 10*time.Second {
		t.Errorf("expected default interval 10s, got %v", w.interval)
	}
}

func TestWatchKeys_EmptyPaths(t *testing.T) {
	client := newTestVaultClient(t, "http://localhost:8200")
	w := NewKVKeyWatcher(client, "secret", 50*time.Millisecond)
	_, err := w.WatchKeys(context.Background(), []string{})
	if err == nil {
		t.Error("expected error for empty paths")
	}
}

func newTestVaultClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	return c
}
