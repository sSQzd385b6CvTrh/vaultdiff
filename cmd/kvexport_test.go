package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func startKVExportServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Query().Get("list") == "true":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"cfg"}},
			})
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"host": "localhost"}},
			})
		}
	}))
}

func TestKVExportCmd_Output(t *testing.T) {
	srv := startKVExportServer(t)
	defer srv.Close()

	os.Setenv("VAULT_ADDR", srv.URL)
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_ADDR")
	defer os.Unsetenv("VAULT_TOKEN")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"kv-export", "myapp", "--mount", "secret"})

	err := rootCmd.Execute()
	if err != nil {
		t.Logf("command output: %s", buf.String())
	}
}

func TestKVExportCmd_MissingArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"kv-export"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing path argument")
	}
}
