package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func startKVGrepServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(body)
	}))
}

func TestKVGrepCmd_Output(t *testing.T) {
	ts := startKVGrepServer(map[string]interface{}{
		"password": "hunter2",
		"username": "alice",
	})
	defer ts.Close()

	os.Setenv("VAULT_ADDR", ts.URL)
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_ADDR")
	defer os.Unsetenv("VAULT_TOKEN")

	root := &cobra.Command{Use: "vaultdiff"}
	var buf bytes.Buffer
	root.SetOut(&buf)

	for _, sub := range rootCmd.Commands() {
		if sub.Use == "kv-grep <path> <pattern>" {
			root.AddCommand(sub)
			break
		}
	}

	root.SetArgs([]string{"kv-grep", "myapp/config", "hunter2"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "password") {
		t.Errorf("expected 'password' in output, got: %s", out)
	}
}

func TestKVGrepCmd_MissingArgs(t *testing.T) {
	root := &cobra.Command{Use: "vaultdiff"}
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "kv-grep <path> <pattern>" {
			root.AddCommand(sub)
			break
		}
	}
	root.SetArgs([]string{"kv-grep", "only-one-arg"})
	if err := root.Execute(); err == nil {
		t.Error("expected error for missing pattern argument")
	}
}
