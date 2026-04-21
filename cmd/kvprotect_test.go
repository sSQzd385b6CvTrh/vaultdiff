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

func startKVProtectServer(t *testing.T, protected string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			body := map[string]interface{}{
				"data": map[string]interface{}{
					"custom_metadata": map[string]interface{}{"protected": protected},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(body)
		case http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestKVProtectCmd_Protect(t *testing.T) {
	srv := startKVProtectServer(t, "false")
	defer srv.Close()
	os.Setenv("VAULT_ADDR", srv.URL)
	defer os.Unsetenv("VAULT_ADDR")
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_TOKEN")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"kvprotect", "myapp/config"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "protected: myapp/config") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestKVProtectCmd_Check(t *testing.T) {
	srv := startKVProtectServer(t, "true")
	defer srv.Close()
	os.Setenv("VAULT_ADDR", srv.URL)
	defer os.Unsetenv("VAULT_ADDR")
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_TOKEN")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"kvprotect", "--check", "myapp/config"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "protected: true") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestKVProtectCmd_MissingArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"kvprotect"})
	if err := rootCmd.Execute(); err == nil {
		t.Error("expected error for missing args")
	}
}
