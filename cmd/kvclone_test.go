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

func startKVCloneServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/myapp/config":
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{"key": "value"},
					},
				})
			}
		case "/v1/secret/data/myapp/config-clone":
			if r.Method == http.MethodPut || r.Method == http.MethodPost {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{"version": json.Number("1")},
				})
			}
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestKVCloneCmd_Output(t *testing.T) {
	srv := startKVCloneServer(t)
	defer srv.Close()

	t.Setenv("VAULT_ADDR", srv.URL)
	t.Setenv("VAULT_TOKEN", "test-token")

	root := &cobra.Command{Use: "vaultdiff"}
	var buf bytes.Buffer
	kvcloneCmd.SetOut(&buf)
	root.AddCommand(kvcloneCmd)
	root.SetArgs([]string{"kvclone", "myapp/config", "myapp/config-clone", "--src-mount", "secret"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKVCloneCmd_MissingArgs(t *testing.T) {
	root := &cobra.Command{Use: "vaultdiff"}
	root.AddCommand(kvcloneCmd)
	root.SetArgs([]string{"kvclone", "only-one-arg"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "accepts 2 arg") {
		t.Errorf("expected arg count error, got: %v", err)
	}
}
