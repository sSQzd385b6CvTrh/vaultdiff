package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kvLintResponse(data map[string]interface{}, version int) []byte {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{
				"version": version,
			},
		},
	}
	b, _ := json.Marshal(body)
	return b
}

func newLintServer(t *testing.T, data map[string]interface{}, version int, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == http.StatusOK {
			w.Write(kvLintResponse(data, version))
		}
	}))
}

func TestKVLint_NoWarnings(t *testing.T) {
	srv := newLintServer(t, map[string]interface{}{
		"api_key": "abc123",
		"db_pass": "secret",
	}, 3, http.StatusOK)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	linter := NewKVLinter(client, "secret")
	results, err := linter.Lint("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no warnings, got %d", len(results))
	}
}

func TestKVLint_EmptyValue(t *testing.T) {
	srv := newLintServer(t, map[string]interface{}{
		"token": "",
	}, 1, http.StatusOK)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	linter := NewKVLinter(client, "secret")
	results, err := linter.Lint("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected empty-value warning")
	}
	if results[0].Warning != "empty value" {
		t.Errorf("unexpected warning: %s", results[0].Warning)
	}
}

func TestKVLint_UppercaseKey(t *testing.T) {
	srv := newLintServer(t, map[string]interface{}{
		"MyKey": "value",
	}, 2, http.StatusOK)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	linter := NewKVLinter(client, "secret")
	results, err := linter.Lint("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var found bool
	for _, r := range results {
		if r.Warning == "key contains uppercase letters" {
			found = true
		}
	}
	if !found {
		t.Error("expected uppercase-key warning")
	}
}

func TestKVLint_NotFound(t *testing.T) {
	srv := newLintServer(t, nil, 0, http.StatusNotFound)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	linter := NewKVLinter(client, "secret")
	_, err := linter.Lint("missing/path")
	if err == nil {
		t.Fatal("expected error for not-found secret")
	}
}

func TestNewKVLinter_DefaultMount(t *testing.T) {
	client := newTestClient(t, "http://127.0.0.1:8200")
	linter := NewKVLinter(client, "")
	if linter.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", linter.mount)
	}
}
