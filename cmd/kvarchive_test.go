package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func startKVArchiveServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/secret/metadata/app/config":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"versions": map[string]interface{}{
						"1": map[string]interface{}{"destroyed": false, "deletion_time": ""},
					},
				},
			})
		case "/v1/secret/data/app/config":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"host": "localhost"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestKVArchiveCmd_Output(t *testing.T) {
	srv := startKVArchiveServer(t)
	defer srv.Close()
	os.Setenv("VAULT_ADDR", srv.URL)
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_ADDR")
	defer os.Unsetenv("VAULT_TOKEN")

	buf := &bytes.Buffer{}
	kvArchiveCmd.SetOut(buf)
	kvArchiveCmd.SetArgs([]string{"app/config", "--mount", "secret"})
	if err := kvArchiveCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "VERSION") {
		t.Errorf("expected VERSION header in output, got: %s", out)
	}
	if !strings.Contains(out, "1") {
		t.Errorf("expected version 1 in output, got: %s", out)
	}
}

func TestKVArchiveCmd_MissingArgs(t *testing.T) {
	kvArchiveCmd.SetArgs([]string{})
	if err := kvArchiveCmd.Execute(); err == nil {
		t.Fatal("expected error for missing path argument")
	}
}
