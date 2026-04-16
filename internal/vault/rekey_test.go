package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func rekeyStatusResponse(started bool, t, n, progress int) map[string]interface{} {
	return map[string]interface{}{
		"started":  started,
		"t":        t,
		"n":        n,
		"progress": progress,
		"required": 2,
		"backup":   false,
	}
}

func newRekeyServer(status int, body interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestRekeyStatus_NotStarted(t *testing.T) {
	srv := newRekeyServer(http.StatusOK, rekeyStatusResponse(false, 0, 0, 0))
	defer srv.Close()

	checker := NewRekeyChecker(srv.URL, "test-token")
	result, err := checker.RekeyStatusResult(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Started {
		t.Error("expected Started=false")
	}
}

func TestRekeyStatus_InProgress(t *testing.T) {
	srv := newRekeyServer(http.StatusOK, rekeyStatusResponse(true, 3, 5, 1))
	defer srv.Close()

	checker := NewRekeyChecker(srv.URL, "test-token")
	result, err := checker.RekeyStatusResult(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Started {
		t.Error("expected Started=true")
	}
	if result.T != 3 || result.N != 5 {
		t.Errorf("unexpected t/n: %d/%d", result.T, result.N)
	}
}

func TestRekeyStatus_UnexpectedStatus(t *testing.T) {
	srv := newRekeyServer(http.StatusForbidden, nil)
	defer srv.Close()

	checker := NewRekeyChecker(srv.URL, "test-token")
	_, err := checker.RekeyStatusResult(context.Background())
	if err == nil {
		t.Fatal("expected error for forbidden response")
	}
}
