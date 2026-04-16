package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func snapshotKVResponse(version int, data map[string]interface{}, destroyed bool) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{
				"version":      float64(version),
				"destroyed":    destroyed,
				"created_time": "2024-01-15T10:00:00.000000000Z",
			},
		},
	}
}

func newSnapshotServer(t *testing.T, status int, body interface{}) (*httptest.Server, *vaultapi.Client) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating vault client: %v", err)
	}
	return ts, client
}

func TestCapture_Success(t *testing.T) {
	body := snapshotKVResponse(3, map[string]interface{}{"api_key": "abc123", "timeout": "30"}, false)
	ts, client := newSnapshotServer(t, 200, body)
	defer ts.Close()

	s := NewSnapshotter(client, "secret")
	snap, err := s.Capture(context.Background(), "myapp/config", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Version != 3 {
		t.Errorf("expected version 3, got %d", snap.Version)
	}
	if snap.Data["api_key"] != "abc123" {
		t.Errorf("expected api_key=abc123, got %q", snap.Data["api_key"])
	}
	if snap.Deleted {
		t.Error("expected Deleted=false")
	}
	if snap.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCapture_NotFound(t *testing.T) {
	ts, client := newSnapshotServer(t, 404, map[string]interface{}{})
	defer ts.Close()

	s := NewSnapshotter(client, "secret")
	_, err := s.Capture(context.Background(), "missing/path", 0)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestCapture_Destroyed(t *testing.T) {
	body := snapshotKVResponse(2, map[string]interface{}{}, true)
	ts, client := newSnapshotServer(t, 200, body)
	defer ts.Close()

	s := NewSnapshotter(client, "secret")
	snap, err := s.Capture(context.Background(), "myapp/config", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !snap.Deleted {
		t.Error("expected Deleted=true for destroyed version")
	}
}

func TestCapture_EmptyData(t *testing.T) {
	body := snapshotKVResponse(1, map[string]interface{}{}, false)
	ts, client := newSnapshotServer(t, 200, body)
	defer ts.Close()

	s := NewSnapshotter(client, "secret")
	snap, err := s.Capture(context.Background(), "myapp/empty", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snap.Data) != 0 {
		t.Errorf("expected empty data map, got %d entries", len(snap.Data))
	}
	if snap.Deleted {
		t.Error("expected Deleted=false for non-destroyed empty secret")
	}
}

func TestNewSnapshotter_DefaultMount(t *testing.T) {
	_, client := newSnapshotServer(t, 200, map[string]interface{}{})
	s := NewSnapshotter(client, "secret")
	if s == nil {
		t.Fatal("expected non-nil Snapshotter")
	}
}

func TestCapture_ServerError(t *testing.T) {
	ts, client := newSnapshotServer(t, 500, map[string]interface{}{"errors": []string{"internal server error"}})
	defer ts.Close()

	s := NewSnapshotter(client, "secret")
	_, err := s.Capture(context.Background(), "myapp/config", 1)
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}
