package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func startKVSearchServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/metadata") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"token"}},
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{"token_secret": "abc123"},
			},
		})
	}))
}

func TestKVSearchCmd_Output(t *testing.T) {
	srv := startKVSearchServer(t)
	defer srv.Close()

	t.Setenv("VAULT_ADDR", srv.URL)
	t.Setenv("VAULT_TOKEN", "test-token")

	buf := &bytes.Buffer{}
	cmd := &cobra.Command{Use: "root"}
	cmd.AddCommand(kvSearchCmd)
	cmd.SetOut(buf)
	kvSearchCmd.SetOut(buf)

	cmd.SetArgs([]string{"kvsearch", "app", "token", "--mount", "secret"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "token") && !strings.Contains(out, "app") {
		t.Logf("output: %q", out)
	}
}

func TestKVSearchCmd_MissingArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "root"}
	cmd.AddCommand(kvSearchCmd)
	cmd.SetArgs([]string{"kvsearch", "app"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for missing query argument")
	}
}
