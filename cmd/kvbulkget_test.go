package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func startKVBulkGetServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/db/creds":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"username": "root", "password": "hunter2"},
				},
			})
		case "/v1/secret/data/app/config":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"env": "production"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestKVBulkGetCmd_Output(t *testing.T) {
	srv := startKVBulkGetServer(t)
	defer srv.Close()

	t.Setenv("VAULT_ADDR", srv.URL)
	t.Setenv("VAULT_TOKEN", "test-token")

	cmd := newTestCmd()
	cmd.SetArgs([]string{"kvbulkget", "--mount", "secret", "db/creds", "app/config"})
	out := &strings.Builder{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "[db/creds]") {
		t.Errorf("expected db/creds in output, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "[app/config]") {
		t.Errorf("expected app/config in output, got: %s", out.String())
	}
}

func TestKVBulkGetCmd_MissingArgs(t *testing.T) {
	cmd := newTestCmd()
	cmd.SetArgs([]string{"kvbulkget"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing args")
	}
}
