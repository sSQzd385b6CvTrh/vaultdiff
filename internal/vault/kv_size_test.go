package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kvSizeResponse(data map[string]interface{}, version int) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{
				"version": version,
			},
		},
	}
}

func newKVSizeServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestKVSize_Success(t *testing.T) {
	payload := kvSizeResponse(map[string]interface{}{
		"username": "admin",
		"password": "s3cr3t",
	}, 3)
	srv := newKVSizeServer(t, http.StatusOK, payload)
	defer srv.Close()

	sizer := NewKVSizer(srv.URL, "test-token", "secret")
	result, err := sizer.Measure("myapp/creds", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.KeyCount != 2 {
		t.Errorf("expected 2 keys, got %d", result.KeyCount)
	}
	if result.Version != 3 {
		t.Errorf("expected version 3, got %d", result.Version)
	}
	// "username"(8)+"admin"(5)+"password"(8)+"s3cr3t"(6) = 27
	if result.TotalBytes != 27 {
		t.Errorf("expected 27 bytes, got %d", result.TotalBytes)
	}
	if result.Path != "myapp/creds" {
		t.Errorf("unexpected path: %s", result.Path)
	}
}

func TestKVSize_NotFound(t *testing.T) {
	srv := newKVSizeServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	sizer := NewKVSizer(srv.URL, "test-token", "")
	_, err := sizer.Measure("missing/path", 0)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestKVSize_UnexpectedStatus(t *testing.T) {
	srv := newKVSizeServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	sizer := NewKVSizer(srv.URL, "test-token", "secret")
	_, err := sizer.Measure("myapp/creds", 1)
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestNewKVSizer_DefaultMount(t *testing.T) {
	s := NewKVSizer("http://localhost:8200", "tok", "")
	if s.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", s.mount)
	}
}
